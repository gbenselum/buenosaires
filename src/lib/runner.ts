import { spawn } from 'node:child_process';
import path from 'node:path';
import { db } from './db';
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
            const output = await new Promise<{ stdout: string; stderr: string }>((resolve, reject) => {
                const child = spawn('bash', [scriptPath], {
                    env: { 
                        ...process.env, 
                        PATH: '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin',
                        DEBIAN_FRONTEND: 'noninteractive' 
                    }
                });

                let stdout = '';
                let stderr = '';

                child.stdout.on('data', (data) => { stdout += data.toString(); });
                child.stderr.on('data', (data) => { stderr += data.toString(); });

                child.on('close', (code) => {
                    if (code === 0) {
                        resolve({ stdout, stderr });
                    } else {
                        reject(new Error(`Process exited with code ${code}\n${stderr}`));
                    }
                });

                child.on('error', (err) => {
                    reject(err);
                });
            });

            const duration = Date.now() - startTime;

            await db.addExecution(taskId, {
                timestamp: new Date().toISOString(),
                status: 'success',
                output: output.stdout || output.stderr,
                duration,
            });

            return { success: true, output: output.stdout };
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
