import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ThemeToggle from '../ThemeToggle';

vi.mock('next-themes', async () => {
  const actual = await vi.importActual('next-themes');
  return {
    ...actual as Record<string, unknown>,
    useTheme: () => ({
      theme: 'dark',
      setTheme: vi.fn(),
    }),
  };
});

describe('ThemeToggle', () => {
  it('renders the toggle button', () => {
    render(<ThemeToggle />);
    const button = screen.getByRole('button');
    expect(button).toBeDefined();
  });
});
