'use client';

import { useState } from 'react';
import { LayoutDashboard, FolderGit2, Settings, Menu, X, Terminal } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import ThemeToggle from './ThemeToggle';

const menuItems = [
  { name: 'Dashboard', icon: LayoutDashboard, href: '/' },
  { name: 'Projects', icon: FolderGit2, href: '/projects' },
  { name: 'Tasks', icon: Terminal, href: '/tasks' },
  { name: 'Settings', icon: Settings, href: '/settings' },
];

export default function Sidebar() {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        className="lg:hidden fixed top-4 left-4 z-50 p-2.5 rounded-xl bg-card/80 backdrop-blur-xl border border-border text-foreground shadow-xl"
        onClick={() => setIsOpen(!isOpen)}
      >
        {isOpen ? <X size={24} /> : <Menu size={24} />}
      </button>

      {/* Sidebar Container */}
      <aside
        className={`fixed top-0 left-0 z-40 h-screen w-64 bg-card/60 backdrop-blur-2xl border-r border-border transition-all duration-300 ease-in-out ${
          isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        }`}
      >
        <div className="flex flex-col h-full p-6">
          {/* Logo */}
          <div className="flex items-center justify-between mb-10 px-2">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-primary flex items-center justify-center shadow-lg shadow-primary/30">
                <span className="text-white font-bold text-lg">B</span>
              </div>
              <span className="text-xl font-bold bg-clip-text text-transparent bg-linear-to-r from-foreground to-foreground/60">
                Buenos Aires
              </span>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1.5">
            {menuItems.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group ${
                    isActive
                      ? 'bg-primary text-white shadow-lg shadow-primary/20'
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                  }`}
                  onClick={() => setIsOpen(false)}
                >
                  <item.icon
                    size={20}
                    className={`transition-colors ${
                      isActive ? 'text-white' : 'group-hover:text-primary'
                    }`}
                  />
                  <span className="font-medium">{item.name}</span>
                </Link>
              );
            })}
          </nav>

          {/* Footer with ThemeToggle */}
          <div className="mt-auto pt-6 border-t border-border flex items-center justify-between gap-3 px-2">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-full bg-linear-to-tr from-indigo-500 to-purple-400 border-2 border-background shadow-md" />
              <div className="flex flex-col">
                <span className="text-sm font-semibold text-foreground">Admin</span>
              </div>
            </div>
            <ThemeToggle />
          </div>
        </div>
      </aside>

      {/* Overlay for mobile */}
      {isOpen && (
        <button
          type="button"
          tabIndex={0}
          className="fixed inset-0 z-30 bg-background/40 backdrop-blur-md lg:hidden w-full h-full border-none cursor-default"
          onClick={() => setIsOpen(false)}
          onKeyDown={(e) => {
            if (e.key === 'Escape' || e.key === 'Enter') {
              setIsOpen(false);
            }
          }}
          aria-label="Close sidebar"
        />
      )}
    </>
  );
}
