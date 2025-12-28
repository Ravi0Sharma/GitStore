/**
 * Single source of truth for route paths
 * All dashboard routes are under /dashboard/*
 */

export const routes = {
  // Public routes
  home: '/',
  signIn: '/signin',
  signUp: '/signup',
  
  // Dashboard routes
  dashboard: '/dashboard',
  dashboardRepo: (repoId: string) => `/dashboard/repo/${repoId}`,
  dashboardRepoIssues: (repoId: string) => `/dashboard/repo/${repoId}/issues`,
  dashboardRepoIssueDetail: (repoId: string, issueId: string) => `/dashboard/repo/${repoId}/issues/${issueId}`,
  dashboardRepoBranches: (repoId: string) => `/dashboard/repo/${repoId}/branches`,
  dashboardRepoMerge: (repoId: string) => `/dashboard/repo/${repoId}/merge`,
} as const;

