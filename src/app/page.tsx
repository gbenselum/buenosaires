'use client';

import { useState, useEffect } from 'react';
import { Activity, CheckCircle2, Clock, XCircle, Terminal as TerminalIcon, Loader2 } from "lucide-react";
import { Project, Task } from '@/lib/db';
import { formatDistanceToNow } from 'date-fns';

export default function Home() {
  const [stats, setStats] = useState({
    totalProjects: 0,
    activeTasks: 0,
    successRate: '0%',
    failedJobs: 0
  });
  const [recentActivity, setRecentActivity] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [projectsRes, tasksRes] = await Promise.all([
          fetch('/api/projects'),
          fetch('/api/tasks')
        ]);
        
        const projects: Project[] = await projectsRes.json();
        const tasks: Task[] = await tasksRes.json();

        // Compute stats
        const totalProjects = projects.length;
        const activeTasks = tasks.filter(t => t.status === 'pending').length;
        const failedJobs = tasks.filter(t => t.status === 'failure').length;
        const successJobs = tasks.filter(t => t.status === 'success').length;
        const totalRuns = tasks.filter(t => t.status !== 'unknown').length;
        const successRate = totalRuns > 0 
          ? `${Math.round((successJobs / totalRuns) * 100)}%` 
          : '0%';

        setStats({ totalProjects, activeTasks, successRate, failedJobs });
        
        // Sort tasks by lastRun for activity feed
        const sortedActivity = [...tasks]
          .filter(t => t.lastRun)
          .sort((a, b) => new Date(b.lastRun!).getTime() - new Date(a.lastRun!).getTime())
          .slice(0, 5);
        
        setRecentActivity(sortedActivity);
      } catch (error) {
        console.error('Failed to fetch dashboard data', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="animate-spin text-primary" size={40} />
      </div>
    );
  }

  return (
    <div className="space-y-10">
      {/* Header */}
      <div className="flex flex-col gap-2">
        <h1 className="text-4xl font-extrabold bg-clip-text text-transparent bg-linear-to-r from-foreground to-foreground/50 tracking-tight">
          Dashboard
        </h1>
        <p className="text-muted-foreground font-medium italic">Real-time GitOps environment status</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Projects", value: stats.totalProjects, icon: Activity, color: "text-blue-500", bg: "bg-blue-500/10" },
          { label: "Active Tasks", value: stats.activeTasks, icon: Clock, color: "text-amber-500", bg: "bg-amber-500/10" },
          { label: "Success Rate", value: stats.successRate, icon: CheckCircle2, color: "text-emerald-500", bg: "bg-emerald-500/10" },
          { label: "Failed Jobs", value: stats.failedJobs, icon: XCircle, color: "text-rose-500", bg: "bg-rose-500/10" },
        ].map((stat, i) => (
          <div
            key={i}
            className="p-6 rounded-3xl bg-card border border-border shadow-sm hover:shadow-md transition-all duration-300 group relative overflow-hidden"
          >
            <div className={`absolute top-0 right-0 w-24 h-24 -mr-8 -mt-8 rounded-full ${stat.bg} blur-2xl group-hover:scale-150 transition-transform duration-500`} />
            
            <div className="flex items-start justify-between mb-4 relative z-10">
              <div className={`p-3 rounded-2xl ${stat.bg} ${stat.color} group-hover:scale-110 transition-transform duration-300 shadow-sm`}>
                <stat.icon size={26} />
              </div>
              <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider bg-accent px-2 py-1 rounded-lg">
                Live
              </span>
            </div>
            <div className="space-y-1 relative z-10">
              <h3 className="text-3xl font-black text-foreground tabular-nums tracking-tight">{stat.value}</h3>
              <p className="text-sm font-semibold text-muted-foreground">{stat.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="rounded-[2.5rem] bg-card border border-border overflow-hidden shadow-sm">
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-foreground">Recent Activity</h2>
            <p className="text-sm text-muted-foreground mt-1">Latest task executions across your projects</p>
          </div>
          <button className="px-5 py-2.5 rounded-xl bg-accent hover:bg-muted text-foreground font-semibold text-sm transition-all shadow-sm">
            View Logs
          </button>
        </div>
        <div className="divide-y divide-border">
          {recentActivity.length > 0 ? (
            recentActivity.map((task) => (
              <div key={task.id} className="p-6 hover:bg-accent/50 transition-colors flex items-center gap-5">
                <div className={`w-12 h-12 rounded-2xl flex items-center justify-center shadow-sm ${
                  task.status === 'success' ? 'bg-emerald-500/10 text-emerald-500' : 'bg-rose-500/10 text-rose-500'
                }`}>
                  <TerminalIcon size={24} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <span className="font-bold text-foreground truncate text-lg tracking-tight">{task.name}</span>
                    <span className={`text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border shadow-xs ${
                      task.status === 'success' 
                        ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' 
                        : 'bg-rose-500/10 text-rose-500 border-rose-500/20'
                    }`}>
                      {task.status}
                    </span>
                  </div>
                  <p className="text-sm font-medium text-muted-foreground truncate">
                    Path: <code className="text-primary font-mono bg-accent px-1.5 py-0.5 rounded text-xs">{task.path}</code>
                  </p>
                </div>
                <span className="text-xs font-bold text-muted-foreground whitespace-nowrap bg-accent px-3 py-1.5 rounded-xl border border-border">
                  {task.lastRun ? formatDistanceToNow(new Date(task.lastRun), { addSuffix: true }) : 'Never'}
                </span>
              </div>
            ))
          ) : (
            <div className="p-12 text-center text-muted-foreground font-medium">
              No recent activity found. Run some tasks to see them here.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
