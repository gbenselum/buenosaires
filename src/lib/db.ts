import fs from 'node:fs/promises';
import path from 'node:path';
import { v4 as uuidv4 } from 'uuid';

const DB_PATH = path.join(process.cwd(), 'db.json');

export interface Project {
    id: string;
    name: string;
    url: string;
    branch: string;
    lastSync?: string;
    status: 'active' | 'error' | 'inactive';
}

export interface Task {
    id: string;
    projectId: string;
    name: string;
    path: string;
    lastRun?: string;
    status: 'success' | 'failure' | 'pending' | 'unknown';
    history: Execution[];
}

export interface Execution {
    id: string;
    timestamp: string;
    status: 'success' | 'failure';
    output: string;
    duration: number;
}

interface Database {
    projects: Project[];
    tasks: Task[];
}

const defaultDb: Database = {
    projects: [],
    tasks: [],
};

async function readDb(): Promise<Database> {
    try {
        const data = await fs.readFile(DB_PATH, 'utf-8');
        return JSON.parse(data);
    } catch {
        // If file doesn't exist, return default
        return defaultDb;
    }
}

async function writeDb(db: Database): Promise<void> {
    await fs.writeFile(DB_PATH, JSON.stringify(db, null, 2));
}

export const db = {
    getProjects: async () => {
        const data = await readDb();
        return data.projects;
    },

    addProject: async (project: Omit<Project, 'id' | 'status'>) => {
        const data = await readDb();
        const newProject: Project = {
            ...project,
            id: uuidv4(),
            status: 'active',
        };
        data.projects.push(newProject);
        await writeDb(data);
        return newProject;
    },

    getProject: async (id: string) => {
        const data = await readDb();
        return data.projects.find((p) => p.id === id);
    },

    getTasks: async () => {
        const data = await readDb();
        return data.tasks;
    },

    getTasksByProject: async (projectId: string) => {
        const data = await readDb();
        return data.tasks.filter(t => t.projectId === projectId);
    },

    upsertTask: async (task: Omit<Task, 'id' | 'history' | 'status'>) => {
        const data = await readDb();
        const existingTaskIndex = data.tasks.findIndex(
            (t) => t.projectId === task.projectId && t.path === task.path
        );

        if (existingTaskIndex >= 0) {
            // Update existing
            const existing = data.tasks[existingTaskIndex];
            data.tasks[existingTaskIndex] = { ...existing, ...task };
            await writeDb(data);
            return data.tasks[existingTaskIndex];
        } else {
            // Create new
            const newTask: Task = {
                ...task,
                id: uuidv4(),
                status: 'unknown',
                history: [],
            };
            data.tasks.push(newTask);
            await writeDb(data);
            return newTask;
        }
    },

    addExecution: async (taskId: string, execution: Omit<Execution, 'id'>) => {
        const data = await readDb();
        const taskIndex = data.tasks.findIndex(t => t.id === taskId);
        if (taskIndex >= 0) {
            const newExecution: Execution = { ...execution, id: uuidv4() };
            data.tasks[taskIndex].history.unshift(newExecution);
            data.tasks[taskIndex].lastRun = execution.timestamp;
            data.tasks[taskIndex].status = execution.status;
            await writeDb(data);
            return newExecution;
        }
        return null;
    }
};
