'use client';

import { useState } from 'react';
import { ThemeToggle } from '@/components/theme-toggle';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/stores/useAuthStore';
import { Menu, X } from 'lucide-react';
import Link from 'next/link';
import { Logo } from '../ui/logo';

export default function LandingHeader() {
  const { user, isAuthenticated } = useAuth();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 w-full border-b border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-950/90 backdrop-blur-xl">
      <div className="container mx-auto px-4 py-3 md:py-4 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center gap-2">
          <Logo size='xl'/>
          <span className="text-lg md:text-xl font-bold text-slate-900 dark:text-white">Seentics</span>
        </div>

        {/* Desktop Nav */}
        <nav className="hidden md:flex items-center gap-8">
          <Link href="#pricing" className="text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition-colors">Pricing</Link>
          <Link href="/docs" className="text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition-colors">Documentation</Link>
          <Link href="/contact" className="text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition-colors">Contact</Link>
        </nav>

        {/* Desktop Actions + Mobile Toggle */}
        <div className="flex items-center gap-3">
          <ThemeToggle />
          {/* Desktop CTAs */}
          <div className="hidden md:flex items-center gap-3">
            {isAuthenticated && user ? (
              <Link href="/websites">
                <Button size="sm">Go to Dashboard</Button>
              </Link>
            ) : (
              <>
                <Link href="/signin">
                  <Button size="sm" variant="outline">Sign In</Button>
                </Link>
                <Link href="/signup">
                  <Button size="sm">Get Started</Button>
                </Link>
              </>
            )}
          </div>
          {/* Mobile menu button */}
          <button
            className="md:hidden inline-flex items-center justify-center rounded-md p-2 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-primary"
            aria-label="Toggle navigation"
            aria-expanded={mobileOpen}
            onClick={() => setMobileOpen(v => !v)}
          >
            {mobileOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>
      </div>

      {/* Mobile Nav Panel */}
      {mobileOpen && (
        <div className="md:hidden border-t border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-950/95 backdrop-blur">
          <div className="container mx-auto px-4 py-3 flex flex-col gap-3">
            <Link href="#pricing" className="py-2 text-sm font-medium text-slate-700 dark:text-slate-200" onClick={() => setMobileOpen(false)}>Pricing</Link>
            <Link href="/docs" className="py-2 text-sm font-medium text-slate-700 dark:text-slate-200" onClick={() => setMobileOpen(false)}>Documentation</Link>
            <Link href="/contact" className="py-2 text-sm font-medium text-slate-700 dark:text-slate-200" onClick={() => setMobileOpen(false)}>Contact</Link>
            <div className="pt-2 flex items-center gap-2">
              {isAuthenticated && user ? (
                <Link href="/websites" onClick={() => setMobileOpen(false)}>
                  <Button className="w-full">Go to Dashboard</Button>
                </Link>
              ) : (
                <>
                  <Link href="/signin" onClick={() => setMobileOpen(false)}>
                    <Button variant="outline" className="w-full">Sign In</Button>
                  </Link>
                  <Link href="/signup" onClick={() => setMobileOpen(false)}>
                    <Button className="w-full">Get Started</Button>
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </header>
  );
}


