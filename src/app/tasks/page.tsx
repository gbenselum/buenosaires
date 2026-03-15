'use client';

import { useState, useEffect } from 'react';
import { Terminal, Clock, CheckCircle2, XCircle, Loader2, AlertCircle, Hash, ChevronRight } from "lucide-react";
import { Task } from '@/lib/db';
import { formatDistanceToNow } from 'date-fns';

export default function TasksPage() {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [loading, setLoading] = useState(true);
    const [runningTaskId, setRunningTaskId] = useState<string | null>(null);

    const fetchTasks = async () => {
        try {
            const res = await fetch('/api/tasks');
            const data = await res.json();
            setTasks(data);
        } catch (error) {
            console.error('Failed to fetch tasks', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTasks();
        const interval = setInterval(fetchTasks, 5000); // Poll every 5 seconds
        return () => clearInterval(interval);
    }, []);

    const handleRunTask = async (taskId: string) => {
        setRunningTaskId(taskId);
        try {
            await fetch('/api/tasks/run', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ taskId }),
            });
            fetchTasks();
        } catch (error) {
            console.error('Failed to run task', error);
        } finally {
            setRunningTaskId(null);
        }
    };

    const getStatusStyles = (status: string) => {
        switch (status) {
            case 'success': return 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20';
            case 'failure': return 'text-rose-500 bg-rose-500/10 border-rose-500/20';
            case 'pending': return 'text-amber-500 bg-amber-500/10 border-amber-500/20';
            default: return 'text-muted-foreground bg-accent/50 border-border';
        }
    };

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'success': return <CheckCircle2 size={16} />;
            case 'failure': return <XCircle size={16} />;
            case 'pending': return <Loader2 size={16} className="animate-spin" />;
            default: return <AlertCircle size={16} />;
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'success': return 'bg-emerald-500';
            case 'failure': return 'bg-rose-500';
            case 'pending': return 'bg-amber-500 animate-pulse';
            default: return 'bg-border';
        }
    };

    return (
        <div className="space-y-10">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-4xl font-extrabold bg-clip-text text-transparent bg-linear-to-r from-foreground to-foreground/50 tracking-tight">
                        Tasks
                    </h1>
                    <p className="text-muted-foreground font-medium mt-1">Monitor and automate your infrastructure scripts</p>
                </div>
                <div className="flex items-center gap-3 bg-card border border-border px-4 py-2 rounded-2xl shadow-sm">
                   <Hash size={18} className="text-primary" />
                   <span className="font-black tabular-nums">{tasks.length}</span>
                   <span className="text-xs font-bold uppercase text-muted-foreground">Scripts</span>
                </div>
            </div>

            {loading ? (
                <div className="flex justify-center py-20">
                    <Loader2 className="animate-spin text-primary" size={48} />
                </div>
            ) : (
                <div className="space-y-4">
                    {tasks.map((task) => (
                        <div 
                            key={task.id} 
                            className="group flex flex-col md:flex-row md:items-center gap-4 p-5 md:p-6 bg-card border border-border rounded-3xl hover:border-primary/40 hover:shadow-lg transition-all duration-300 relative overflow-hidden"
                        >
                            {/* Status Accent Bar */}
                            <div className={`absolute left-0 top-0 bottom-0 w-1.5 ${getStatusColor(task.status)}`} />

                            <div className="flex items-center gap-4 flex-1 min-w-0">
                                <div className="p-3.5 rounded-2xl bg-accent text-primary group-hover:scale-110 transition-transform duration-300 border border-border shrink-0 shadow-xs">
                                    <Terminal size={24} />
                                </div>
                                <div className="min-w-0 flex-1">
                                    <h3 className="text-lg font-black text-foreground truncate tracking-tight">{task.name}</h3>
                                    <div className="flex items-center gap-2 mt-0.5">
                                        <code className="text-[10px] font-bold text-muted-foreground bg-accent/50 px-2 py-0.5 rounded-lg border border-border italic truncate">
                                            {task.path}
                                        </code>
                                    </div>
                                </div>
                            </div>

                            <div className="flex flex-wrap items-center gap-4 md:gap-8">
                                <div className="flex flex-col gap-1 min-w-[120px]">
                                    <span className="text-[10px] uppercase font-black tracking-widest text-muted-foreground">Last Execution</span>
                                    <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                                        <Clock size={14} className="text-muted-foreground" />
                                        <span>
                                            {task.lastRun
                                                ? formatDistanceToNow(new Date(task.lastRun), { addSuffix: true })
                                                : 'Never'}
                                        </span>
                                    </div>
                                </div>

                                <div className="flex flex-col gap-1">
                                    <span className="text-[10px] uppercase font-black tracking-widest text-muted-foreground">Status</span>
                                    <span className={`inline-flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-black uppercase tracking-widest border shadow-xs ${getStatusStyles(task.status)}`}>
                                        {getStatusIcon(task.status)}
                                        {task.status}
                                    </span>
                                </div>

                                <div className="flex items-center gap-2 ml-auto">
                                    <button
                                        onClick={() => handleRunTask(task.id)}
                                        disabled={runningTaskId === task.id || task.status === 'pending'}
                                        className="inline-flex items-center justify-center p-3.5 rounded-2xl bg-primary text-white hover:bg-primary/90 transition-all shadow-lg shadow-primary/20 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed group/btn overflow-hidden relative"
                                    >
                                        {runningTaskId === task.id ? (
                                            <Loader2 size={24} className="animate-spin" />
                                        ) : (
                                            <div className="flex items-center gap-2 font-black uppercase tracking-widest text-xs">
                                                <span>Run Script</span>
                                                <ChevronRight size={18} className="translate-x-0 group-hover/btn:translate-x-1 transition-transform" />
                                            </div>
                                        )}
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}

                    {tasks.length === 0 && (
                        <div className="py-20 text-center rounded-[2.5rem] border-2 border-dashed border-border bg-accent/30">
                            <Terminal className="mx-auto text-muted-foreground mb-4 opacity-50" size={64} />
                            <h3 className="text-xl font-bold text-foreground font-mono">_no_tasks_found</h3>
                            <p className="text-muted-foreground font-medium mt-1">Add a project with .sh files for them to appear here</p>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
