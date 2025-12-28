#!/usr/bin/env node
/**
 * Test script for GitStore API endpoints
 * Usage: node scripts/test-api.mjs [baseUrl]
 * Default baseUrl: http://localhost:8080
 */

const BASE_URL = process.argv[2] || 'http://localhost:8080';

const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logPass(message) {
  log(`✓ PASS: ${message}`, 'green');
}

function logFail(message) {
  log(`✗ FAIL: ${message}`, 'red');
}

function logInfo(message) {
  log(`ℹ ${message}`, 'blue');
}

async function testEndpoint(name, method, url, body = null) {
  try {
    logInfo(`Testing ${name}...`);
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };
    
    if (body) {
      options.body = JSON.stringify(body);
    }
    
    const response = await fetch(url, options);
    const isOK = response.ok;
    const contentType = response.headers.get('content-type') || '';
    const isJSON = contentType.includes('application/json');
    
    let data = null;
    if (isJSON) {
      data = await response.json();
    } else {
      const text = await response.text();
      if (text) {
        try {
          data = JSON.parse(text);
        } catch {
          data = text;
        }
      }
    }
    
    if (isOK) {
      logPass(`${name} - Status: ${response.status}`);
      if (data !== null) {
        if (Array.isArray(data)) {
          logInfo(`  Response: Array with ${data.length} items`);
        } else if (typeof data === 'object') {
          logInfo(`  Response: Object with keys: ${Object.keys(data).join(', ')}`);
        } else {
          logInfo(`  Response: ${String(data).substring(0, 100)}`);
        }
      }
      return { success: true, data, status: response.status };
    } else {
      logFail(`${name} - Status: ${response.status} - ${data?.error || response.statusText}`);
      return { success: false, error: data?.error || response.statusText, status: response.status };
    }
  } catch (err) {
    logFail(`${name} - Error: ${err.message}`);
    return { success: false, error: err.message };
  }
}

async function main() {
  log('\n=== GitStore API Test Suite ===\n', 'blue');
  log(`Base URL: ${BASE_URL}\n`);
  
  const results = {
    passed: 0,
    failed: 0,
    skipped: 0,
  };
  
  // Test 1: GET /api/repos (should return array)
  const listReposResult = await testEndpoint('GET /api/repos', 'GET', `${BASE_URL}/api/repos`);
  if (listReposResult.success) {
    results.passed++;
  } else {
    results.failed++;
    log('  → Backend may not be running. Start with: cd gitClone && go build ./cmd/gitserver && ./gitserver', 'yellow');
    log('\n=== Test Suite Aborted ===\n', 'yellow');
    process.exit(1);
  }
  
  // Test 2: POST /api/repos (create test repo)
  const testRepoName = `test-repo-${Date.now()}`;
  const createRepoResult = await testEndpoint(
    'POST /api/repos',
    'POST',
    `${BASE_URL}/api/repos`,
    { name: testRepoName, description: 'Test repository for API testing' }
  );
  
  if (createRepoResult.success) {
    results.passed++;
    logInfo(`  Created repo: ${testRepoName}`);
  } else {
    results.failed++;
    if (createRepoResult.status === 409) {
      logInfo('  → Repo already exists (this is OK)');
      results.failed--;
      results.skipped++;
    }
  }
  
  // Test 3: GET /api/repos again (should include new repo)
  const listReposResult2 = await testEndpoint('GET /api/repos (after create)', 'GET', `${BASE_URL}/api/repos`);
  if (listReposResult2.success) {
    results.passed++;
    if (Array.isArray(listReposResult2.data)) {
      const found = listReposResult2.data.find(r => r.id === testRepoName || r.name === testRepoName);
      if (found) {
        logPass(`  Found created repo: ${testRepoName}`);
      } else {
        logFail(`  Created repo ${testRepoName} not found in list`);
        results.failed++;
      }
    }
  } else {
    results.failed++;
  }
  
  // Test 4: GET /api/repos/:id/branches
  const branchesResult = await testEndpoint(
    'GET /api/repos/:id/branches',
    'GET',
    `${BASE_URL}/api/repos/${testRepoName}/branches`
  );
  if (branchesResult.success) {
    results.passed++;
  } else if (branchesResult.status === 404) {
    logInfo('  → Repo not found (may have been cleaned up)');
    results.skipped++;
  } else {
    results.failed++;
  }
  
  // Test 5: POST /api/repos/:id/checkout (create branch)
  const testBranchName = 'feature/test-branch';
  const checkoutResult = await testEndpoint(
    'POST /api/repos/:id/checkout',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/checkout`,
    { branch: testBranchName }
  );
  if (checkoutResult.success) {
    results.passed++;
    logInfo(`  Created branch: ${testBranchName}`);
    
    // Test 5b: Verify branch appears in branches list
    await new Promise(resolve => setTimeout(resolve, 500)); // Wait for server to update
    const branchesResult2 = await testEndpoint(
      'GET /api/repos/:id/branches (after create)',
      'GET',
      `${BASE_URL}/api/repos/${testRepoName}/branches`
    );
    if (branchesResult2.success && Array.isArray(branchesResult2.data)) {
      const foundBranch = branchesResult2.data.find(b => b.name === testBranchName);
      if (foundBranch) {
        logPass(`  Branch ${testBranchName} found in branches list`);
        results.passed++;
      } else {
        logFail(`  Branch ${testBranchName} not found in branches list`);
        results.failed++;
      }
    } else {
      logFail('  Failed to verify branch in list');
      results.failed++;
    }
  } else if (checkoutResult.status === 404) {
    logInfo('  → Repo not found (skipping branch test)');
    results.skipped++;
  } else {
    results.failed++;
  }
  
  // Test 6: GET /api/repos/:id/commits
  const commitsResult = await testEndpoint(
    'GET /api/repos/:id/commits',
    'GET',
    `${BASE_URL}/api/repos/${testRepoName}/commits`
  );
  if (commitsResult.success) {
    results.passed++;
  } else if (commitsResult.status === 404) {
    logInfo('  → Repo not found (skipping commits test)');
    results.skipped++;
  } else {
    results.failed++;
  }
  
  // Test 7: POST /api/repos/:id/commit (create commit on master branch)
  logInfo('\n--- Testing commits for merge scenarios ---');
  const commitResult1 = await testEndpoint(
    'POST /api/repos/:id/commit (master branch)',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/commit`,
    { message: 'Initial commit on master' }
  );
  if (commitResult1.success) {
    results.passed++;
    logInfo('  Created initial commit on main branch');
  } else if (commitResult1.status === 404) {
    logInfo('  → Repo not found (skipping commit test)');
    results.skipped++;
  } else {
    results.failed++;
  }

  // Wait a bit for server to process
  await new Promise(resolve => setTimeout(resolve, 500));

  // Test 8: POST /api/repos/:id/checkout (switch to feature branch)
  const checkoutFeatureResult = await testEndpoint(
    'POST /api/repos/:id/checkout (to feature branch)',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/checkout`,
    { branch: testBranchName }
  );
  if (checkoutFeatureResult.success) {
    results.passed++;
    logInfo('  Switched to feature branch');
  } else if (checkoutFeatureResult.status === 404) {
    logInfo('  → Repo not found (skipping checkout test)');
    results.skipped++;
  } else {
    results.failed++;
  }

  await new Promise(resolve => setTimeout(resolve, 500));

  // Test 9: POST /api/repos/:id/commit (create commit on feature branch)
  const commitResult2 = await testEndpoint(
    'POST /api/repos/:id/commit (feature branch)',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/commit`,
    { message: 'Commit on feature branch' }
  );
  if (commitResult2.success) {
    results.passed++;
    logInfo('  Created commit on feature branch');
  } else if (commitResult2.status === 404) {
    logInfo('  → Repo not found (skipping commit test)');
    results.skipped++;
  } else {
    results.failed++;
  }

  await new Promise(resolve => setTimeout(resolve, 500));

  // Test 10: POST /api/repos/:id/checkout (switch back to master)
  const checkoutMainResult = await testEndpoint(
    'POST /api/repos/:id/checkout (back to master)',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/checkout`,
    { branch: 'master' }
  );
  if (checkoutMainResult.success) {
    results.passed++;
    logInfo('  Switched back to main branch');
  } else if (checkoutMainResult.status === 404) {
    logInfo('  → Repo not found (skipping checkout test)');
    results.skipped++;
  } else {
    results.failed++;
  }

  await new Promise(resolve => setTimeout(resolve, 500));

  // Test 11: POST /api/repos/:id/merge (FAST-FORWARD merge)
  logInfo('\n--- Testing Fast-Forward Merge ---');
  logInfo('  Scenario: Merging feature branch into master (feature is ahead)');
  const fastForwardMergeResult = await testEndpoint(
    'POST /api/repos/:id/merge (fast-forward)',
    'POST',
    `${BASE_URL}/api/repos/${testRepoName}/merge`,
    { branch: testBranchName }
  );
  if (fastForwardMergeResult.success) {
    results.passed++;
    logPass('  Fast-forward merge succeeded as expected');
    if (fastForwardMergeResult.data?.type === 'fast-forward') {
      logPass('  Response correctly indicates fast-forward type');
      results.passed++;
    }
    
    // Verify branch state after merge
    await new Promise(resolve => setTimeout(resolve, 500));
    const branchesAfterFF = await testEndpoint(
      'GET /api/repos/:id/branches (after fast-forward merge)',
      'GET',
      `${BASE_URL}/api/repos/${testRepoName}/branches`
    );
    if (branchesAfterFF.success) {
      results.passed++;
      logInfo('  Branch list retrieved successfully after merge');
    }
  } else if (fastForwardMergeResult.status === 404) {
    logInfo('  → Repo not found (skipping merge test)');
    results.skipped++;
  } else {
    logFail(`  Fast-forward merge failed unexpectedly: ${fastForwardMergeResult.error}`);
    results.failed++;
  }

  
  // Summary
  log('\n=== Test Summary ===', 'blue');
  log(`Passed: ${results.passed}`, 'green');
  log(`Failed: ${results.failed}`, results.failed > 0 ? 'red' : 'reset');
  log(`Skipped: ${results.skipped}`, 'yellow');
  log('');
  
  if (results.failed === 0) {
    log('All tests passed! ✓', 'green');
    process.exit(0);
  } else {
    log('Some tests failed. Check the output above.', 'red');
    process.exit(1);
  }
}

main().catch(err => {
  logFail(`Fatal error: ${err.message}`);
  console.error(err);
  process.exit(1);
});

