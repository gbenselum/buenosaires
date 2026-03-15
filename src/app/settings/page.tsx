'use client';

import { useTheme } from 'next-themes';
import { Settings, Moon, Sun, Monitor, Bell, Database, Github } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function SettingsPage() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMounted(true);
  }, []);

  if (!mounted) return null;

  return (
    <div className="space-y-10">
      <div>
        <h1 className="text-4xl font-extrabold bg-clip-text text-transparent bg-linear-to-r from-foreground to-foreground/50 tracking-tight">
          Settings
        </h1>
        <p className="text-muted-foreground font-medium mt-1">Configure your workspace preferences</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
        <div className="lg:col-span-2 space-y-8">
            {/* Appearance Section */}
            <section className="bg-card border border-border rounded-[2.5rem] p-8 shadow-sm">
                <div className="flex items-center gap-3 mb-8">
                    <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                        <Monitor size={22} />
                    </div>
                    <h2 className="text-xl font-bold text-foreground">Appearance</h2>
                </div>

                <div className="space-y-6">
                    <div className="flex items-center justify-between p-4 bg-accent/30 rounded-2xl border border-border/50">
                        <div>
                            <p className="font-bold text-foreground">Interface Theme</p>
                            <p className="text-xs text-muted-foreground font-medium">Customize how Buenos Aires looks on your screen</p>
                        </div>
                        <div className="flex bg-accent p-1.5 rounded-[1.25rem] border border-border shadow-inner">
                            {[
                                { id: 'light', icon: Sun, label: 'Light' },
                                { id: 'dark', icon: Moon, label: 'Dark' },
                                { id: 'system', icon: Monitor, label: 'System' }
                            ].map((mode) => (
                                <button
                                    key={mode.id}
                                    onClick={() => setTheme(mode.id)}
                                    className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-black uppercase tracking-wider transition-all ${
                                        theme === mode.id 
                                            ? 'bg-primary text-white shadow-lg shadow-primary/20' 
                                            : 'text-muted-foreground hover:text-foreground'
                                    }`}
                                >
                                    <mode.icon size={14} />
                                    {mode.label}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            </section>

            {/* Notifications Section (Placeholder) */}
            <section className="bg-card border border-border rounded-[2.5rem] p-8 shadow-sm opacity-60">
                <div className="flex items-center gap-3 mb-8">
                    <div className="p-2.5 rounded-xl bg-accent text-muted-foreground">
                        <Bell size={22} />
                    </div>
                    <h2 className="text-xl font-bold text-foreground">Notifications</h2>
                </div>
                <div className="text-center py-4 italic text-sm text-muted-foreground font-medium">
                    Notification settings are coming soon in the next update.
                </div>
            </section>
        </div>

        <div className="space-y-8">
            <div className="p-8 rounded-[2rem] bg-card border border-border shadow-sm">
                <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">System Health</h3>
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 text-foreground font-bold text-sm">
                            <Database size={16} className="text-primary" />
                            <span>Database</span>
                        </div>
                        <span className="text-[10px] font-black uppercase px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">Operational</span>
                    </div>
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 text-foreground font-bold text-sm">
                            <Github size={16} className="text-indigo-500" />
                            <span>Git Gateway</span>
                        </div>
                        <span className="text-[10px] font-black uppercase px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">Active</span>
                    </div>
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 text-foreground font-bold text-sm">
                            <Settings size={16} className="text-muted-foreground" />
                            <span>Version</span>
                        </div>
                        <span className="text-xs font-mono font-bold text-muted-foreground">v0.1.0-alpha</span>
                    </div>
                </div>
            </div>
        </div>
      </div>
    </div>
  );
}
