import { NextResponse } from 'next/server';
import { db } from '@/lib/db';

export async function GET(request: Request) {
    const { searchParams } = new URL(request.url);
    const projectId = searchParams.get('projectId');

    if (projectId) {
        const tasks = await db.getTasksByProject(projectId);
        return NextResponse.json(tasks);
    }

    const tasks = await db.getTasks();
    return NextResponse.json(tasks);
}
