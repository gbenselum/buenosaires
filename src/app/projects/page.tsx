'use client';

import { useState, useEffect } from 'react';
import { FolderGit2, MoreVertical, Plus, Loader2, RefreshCw, ExternalLink, GitBranch, ShieldCheck } from "lucide-react";
import { Project } from '@/lib/db';

export default function ProjectsPage() {
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [newProject, setNewProject] = useState({ name: '', url: '', branch: 'main' });
    const [submitting, setSubmitting] = useState(false);

    const fetchProjects = async () => {
        try {
            const res = await fetch('/api/projects');
            const data = await res.json();
            setProjects(data);
        } catch (error) {
            console.error('Failed to fetch projects', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchProjects();
    }, []);

    const handleAddProject = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        try {
            await fetch('/api/projects', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newProject),
            });
            setNewProject({ name: '', url: '', branch: 'main' });
            setIsModalOpen(false);
            fetchProjects();
        } catch (error) {
            console.error('Failed to add project', error);
        } finally {
            setSubmitting(false);
        }
    };

    const handleSync = async (id: string, e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        try {
            await fetch(`/api/projects/${id}/sync`, { method: 'POST' });
            fetchProjects();
        } catch (error) {
            console.error("Sync failed", error);
        }
    }

    return (
        <div className="space-y-10">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-4xl font-extrabold bg-clip-text text-transparent bg-linear-to-r from-foreground to-foreground/50 tracking-tight">
                        Projects
                    </h1>
                    <p className="text-muted-foreground font-medium mt-1">Manage and monitor your infrastructure repositories</p>
                </div>
                <button
                    onClick={() => setIsModalOpen(true)}
                    className="flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary/90 text-white rounded-2xl transition-all font-bold shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98]"
                >
                    <Plus size={22} strokeWidth={3} />
                    Add Project
                </button>
            </div>

            {loading ? (
                <div className="flex justify-center py-20">
                    <Loader2 className="animate-spin text-primary" size={48} />
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                    {projects.map((project) => (
                        <div
                            key={project.id}
                            className="group relative p-8 rounded-[2rem] bg-card border border-border shadow-sm hover:shadow-xl hover:border-primary/30 transition-all duration-500 overflow-hidden"
                        >
                            <div className="absolute top-0 right-0 w-32 h-32 -mr-12 -mt-12 bg-primary/5 rounded-full blur-3xl group-hover:bg-primary/10 transition-colors duration-500" />
                            
                            <div className="flex justify-between items-start mb-8 relative z-10">
                                <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 group-hover:rotate-3 transition-transform duration-500 shadow-sm border border-primary/10">
                                    <FolderGit2 size={28} />
                                </div>
                                <div className="flex gap-2">
                                    <button 
                                        onClick={(e) => handleSync(project.id, e)} 
                                        className="p-2.5 bg-accent hover:bg-muted rounded-xl text-muted-foreground hover:text-primary transition-all shadow-sm border border-border" 
                                        title="Sync Now"
                                    >
                                        <RefreshCw size={18} />
                                    </button>
                                    <button className="p-2.5 bg-accent hover:bg-muted rounded-xl text-muted-foreground hover:text-foreground transition-all shadow-sm border border-border">
                                        <MoreVertical size={18} />
                                    </button>
                                </div>
                            </div>

                            <div className="mb-8 relative z-10">
                                <h3 className="text-2xl font-black text-foreground mb-2 group-hover:text-primary transition-colors duration-300 tracking-tight leading-none">
                                    {project.name}
                                </h3>
                                <div className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors max-w-full">
                                    <ExternalLink size={14} className="shrink-0" />
                                    <p className="text-sm font-medium truncate italic">{project.url}</p>
                                </div>
                            </div>

                            <div className="space-y-4 pt-6 border-t border-border relative z-10">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <GitBranch size={16} />
                                        <span className="text-xs font-bold uppercase tracking-wider">Branch</span>
                                    </div>
                                    <span className="text-xs font-black font-mono bg-accent text-foreground px-3 py-1 rounded-lg border border-border shadow-xs">
                                        {project.branch}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <ShieldCheck size={16} />
                                        <span className="text-xs font-bold uppercase tracking-wider">Status</span>
                                    </div>
                                    <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-black uppercase tracking-widest border shadow-xs ${
                                        project.status === 'active' 
                                            ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' 
                                            : 'bg-rose-500/10 text-rose-500 border-rose-500/20'
                                    }`}>
                                        <span className={`w-1.5 h-1.5 rounded-full ${project.status === 'active' ? 'bg-emerald-500 animate-pulse' : 'bg-rose-500'}`} />
                                        {project.status}
                                    </span>
                                </div>
                            </div>
                        </div>
                    ))}
                    
                    {projects.length === 0 && (
                        <div className="col-span-full py-20 text-center rounded-[2rem] border-2 border-dashed border-border bg-accent/30">
                            <FolderGit2 className="mx-auto text-muted-foreground mb-4 opacity-50" size={64} />
                            <h3 className="text-xl font-bold text-foreground">No projects found</h3>
                            <p className="text-muted-foreground font-medium mt-1">Start by adding your first GitOps repository</p>
                        </div>
                    )}
                </div>
            )}

            {/* Add Project Modal */}
            {isModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-6 sm:p-4">
                    <div 
                        className="fixed inset-0 bg-background/40 backdrop-blur-md" 
                        onClick={() => setIsModalOpen(false)} 
                    />
                    <div className="w-full max-w-lg bg-card border border-border rounded-[2.5rem] p-10 shadow-3xl relative z-10 animate-in fade-in zoom-in duration-300">
                        <div className="flex justify-between items-center mb-8">
                            <h2 className="text-3xl font-black text-foreground tracking-tight">Add New Project</h2>
                            <button 
                                onClick={() => setIsModalOpen(false)}
                                className="p-2 hover:bg-accent rounded-full transition-colors text-muted-foreground hover:text-foreground"
                            >
                                <Plus size={24} className="rotate-45" />
                            </button>
                        </div>
                        
                        <form onSubmit={handleAddProject} className="space-y-6">
                            <div className="space-y-2">
                                <label className="text-xs font-black uppercase tracking-widest text-muted-foreground ml-1">Project Name</label>
                                <input
                                    type="text"
                                    required
                                    placeholder="e.g. Production Cluster"
                                    className="w-full bg-accent/50 border border-border rounded-2xl px-5 py-4 text-foreground font-bold focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all placeholder:font-medium placeholder:italic"
                                    value={newProject.name}
                                    onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-xs font-black uppercase tracking-widest text-muted-foreground ml-1">Repository URL</label>
                                <input
                                    type="text"
                                    required
                                    placeholder="https://github.com/..."
                                    className="w-full bg-accent/50 border border-border rounded-2xl px-5 py-4 text-foreground font-bold focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all placeholder:font-medium placeholder:italic"
                                    value={newProject.url}
                                    onChange={(e) => setNewProject({ ...newProject, url: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-xs font-black uppercase tracking-widest text-muted-foreground ml-1">Default Branch</label>
                                <input
                                    type="text"
                                    required
                                    placeholder="main"
                                    className="w-full bg-accent/50 border border-border rounded-2xl px-5 py-4 text-foreground font-bold focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all placeholder:font-medium placeholder:italic"
                                    value={newProject.branch}
                                    onChange={(e) => setNewProject({ ...newProject, branch: e.target.value })}
                                />
                            </div>
                            <div className="flex gap-4 mt-10">
                                <button
                                    type="button"
                                    onClick={() => setIsModalOpen(false)}
                                    className="flex-1 px-6 py-4 rounded-2xl bg-accent hover:bg-muted text-foreground font-bold transition-all border border-border"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={submitting}
                                    className="flex-[2] px-6 py-4 bg-primary hover:bg-primary/90 text-white rounded-2xl transition-all font-bold shadow-lg shadow-primary/20 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {submitting ? (
                                        <div className="flex items-center justify-center gap-2">
                                            <Loader2 size={20} className="animate-spin" />
                                            <span>Joining...</span>
                                        </div>
                                    ) : 'Create Project'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
