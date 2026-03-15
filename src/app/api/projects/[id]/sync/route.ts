import { NextResponse } from 'next/server';
import { gitService } from '@/lib/git';

export async function POST(
    request: Request,
    { params }: { params: Promise<{ id: string }> }
) {
    try {
        const { id } = await params;
        const result = await gitService.syncProject(id);
        return NextResponse.json(result);
    } catch {
        return NextResponse.json(
            { error: 'Failed to sync project' },
            { status: 500 }
        );
    }
}
