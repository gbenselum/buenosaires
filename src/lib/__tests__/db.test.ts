import { describe, it, expect, vi, beforeEach } from 'vitest';
import fs from 'node:fs/promises';
import { db } from '../db';

vi.mock('node:fs/promises', () => ({
    default: {
        readFile: vi.fn(),
        writeFile: vi.fn(),
        access: vi.fn(),
        mkdir: vi.fn(),
    },
    readFile: vi.fn(),
    writeFile: vi.fn(),
    access: vi.fn(),
    mkdir: vi.fn(),
}));

describe('db service', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should return empty projects when DB is empty', async () => {
        vi.mocked(fs.readFile).mockRejectedValue(new Error('File not found'));
        const projects = await db.getProjects();
        expect(projects).toEqual([]);
    });

    it('should add a project', async () => {
        vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify({ projects: [], tasks: [] }));
        vi.mocked(fs.writeFile).mockResolvedValue(undefined);

        const newProject = await db.addProject({
            name: 'Test Project',
            url: 'https://github.com/test/repo',
            branch: 'main'
        });

        expect(newProject.name).toBe('Test Project');
        expect(newProject.id).toBeDefined();
        expect(newProject.status).toBe('active');
        expect(fs.writeFile).toHaveBeenCalled();
    });
});
