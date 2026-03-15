import { exec } from 'child_process';
import path from 'path';
import { promisify } from 'util';
import { db } from './db';

const execAsync = promisify(exec);
const REPOS_DIR = path.join(process.cwd(), 'repos');

export const runnerService = {
    runTask: async (taskId: string) => {
        const tasks = await db.getTasks();
        const task = tasks.find((t) => t.id === taskId);
        if (!task) throw new Error('Task not found');

        const project = await db.getProject(task.projectId);
        if (!project) throw new Error('Project not found');

        const scriptPath = path.join(REPOS_DIR, project.id, task.path);
        const startTime = Date.now();

        try {
            const { stdout, stderr } = await execAsync(`bash "${scriptPath}"`);
            const duration = Date.now() - startTime;

            await db.addExecution(taskId, {
                timestamp: new Date().toISOString(),
                status: 'success',
                output: stdout || stderr,
                duration,
            });

            return { success: true, output: stdout };
        } catch (error: unknown) {
            const duration = Date.now() - startTime;
            const message = error instanceof Error ? error.message : String(error);
            await db.addExecution(taskId, {
                timestamp: new Date().toISOString(),
                status: 'failure',
                output: message,
                duration,
            });
            throw error;
        }
    },
};
