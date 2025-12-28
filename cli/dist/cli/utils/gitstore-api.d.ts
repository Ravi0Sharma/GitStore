/**
 * GitStore API client for CLI
 * Reuses the API structure from Client/src/lib/api.ts
 */
export interface RepoListItem {
    id: string;
    name: string;
    description?: string;
    currentBranch: string;
    branchCount: number;
    commitCount: number;
    lastUpdated?: string;
}
export interface Branch {
    name: string;
    createdAt: string;
}
export interface Commit {
    hash: string;
    message: string;
    author: string;
    date: string;
}
/**
 * Check if GitStore API server is available
 */
export declare function isApiAvailable(): Promise<boolean>;
/**
 * Find repository ID by path
 * This searches through all repos to find one matching the path
 */
export declare function findRepoByPath(repoPath: string): Promise<string | null>;
export declare const gitstoreApi: {
    listRepos(): Promise<RepoListItem[]>;
    createRepo(name: string, description?: string): Promise<RepoListItem>;
    getBranches(repoId: string): Promise<Branch[]>;
    getCommits(repoId: string): Promise<Commit[]>;
    checkout(repoId: string, branch: string): Promise<void>;
    commit(repoId: string, message: string): Promise<void>;
    merge(repoId: string, branch: string): Promise<{
        message: string;
        type?: string;
    }>;
};
export declare const listRepos: () => Promise<RepoListItem[]>, createRepo: (name: string, description?: string) => Promise<RepoListItem>, getBranches: (repoId: string) => Promise<Branch[]>, getCommits: (repoId: string) => Promise<Commit[]>, checkout: (repoId: string, branch: string) => Promise<void>, commit: (repoId: string, message: string) => Promise<void>, merge: (repoId: string, branch: string) => Promise<{
    message: string;
    type?: string;
}>;
//# sourceMappingURL=gitstore-api.d.ts.map