import { NextResponse } from 'next/server';
import { db } from '@/lib/db';
import { gitService } from '@/lib/git';

export async function GET() {
    const projects = await db.getProjects();
    return NextResponse.json(projects);
}

export async function POST(request: Request) {
    try {
        const body = await request.json();
        const { name, url, branch } = body;

        if (!name || !url || !branch) {
            return NextResponse.json(
                { error: 'Missing required fields' },
                { status: 400 }
            );
        }

        const project = await db.addProject({ name, url, branch });

        // Initial sync
        try {
            await gitService.syncProject(project.id);
        } catch (e) {
            console.error("Initial sync failed", e);
        }

        return NextResponse.json(project);
    } catch {
        return NextResponse.json(
            { error: 'Failed to create project' },
            { status: 500 }
        );
    }
}
