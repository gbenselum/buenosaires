import simpleGit from 'simple-git';
import path from 'path';
import fs from 'fs/promises';
import { db } from './db';

const REPOS_DIR = path.join(process.cwd(), 'repos');

// Ensure repos directory exists
async function ensureReposDir() {
    try {
        await fs.access(REPOS_DIR);
    } catch {
        await fs.mkdir(REPOS_DIR, { recursive: true });
    }
}

export const gitService = {
    syncProject: async (projectId: string) => {
        await ensureReposDir();
        const project = await db.getProject(projectId);
        if (!project) throw new Error('Project not found');

        const projectPath = path.join(REPOS_DIR, project.id);
        const git = simpleGit();

        try {
            // Check if repo exists
            try {
                await fs.access(projectPath);
                // Pull
                await simpleGit(projectPath).pull();
            } catch {
                // Clone
                await git.clone(project.url, projectPath);
            }

            // Scan for .sh files
            const files = await findShellScripts(projectPath);

            // Update tasks in DB
            for (const file of files) {
                const relativePath = path.relative(projectPath, file);
                await db.upsertTask({
                    projectId: project.id,
                    name: path.basename(file),
                    path: relativePath,
                });
            }

            return { success: true, files };
        } catch (error) {
            console.error('Git sync error:', error);
            throw error;
        }
    },
};

async function findShellScripts(dir: string): Promise<string[]> {
    const results: string[] = [];
    const entries = await fs.readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        if (entry.isDirectory() && entry.name !== '.git') {
            results.push(...(await findShellScripts(fullPath)));
        } else if (entry.isFile() && entry.name.endsWith('.sh')) {
            results.push(fullPath);
        }
    }

    return results;
}
