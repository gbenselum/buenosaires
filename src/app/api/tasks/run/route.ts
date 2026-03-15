import { NextResponse } from 'next/server';
import { runnerService } from '@/lib/runner';

export async function POST(request: Request) {
    try {
        const body = await request.json();
        const { taskId } = body;

        if (!taskId) {
            return NextResponse.json(
                { error: 'Missing taskId' },
                { status: 400 }
            );
        }

        const result = await runnerService.runTask(taskId);
        return NextResponse.json(result);
    } catch (error: unknown) {
        const message = error instanceof Error ? error.message : 'Failed to run task';
        return NextResponse.json(
            { error: message },
            { status: 500 }
        );
    }
}
