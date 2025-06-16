import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import RootLayout from './layout';

// Mock Providers component
jest.mock('@/components/providers', () => ({
  Providers: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// Mock next/link
jest.mock('next/link', () => {
  return function MockLink({ children, href, ...props }: {
    children: React.ReactNode;
    href: string;
    [key: string]: unknown;
  }) {
    return <a href={href} {...props}>{children}</a>;
  };
});

describe('RootLayout Navigation', () => {
  beforeEach(() => {
    // Set up default viewport size
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    });
    Object.defineProperty(window, 'innerHeight', {
      writable: true,
      configurable: true,
      value: 768,
    });
  });

  it('renders navigation with icons and labels', () => {
    render(
      <RootLayout>
        <div>Test content</div>
      </RootLayout>
    );

    // Check that navigation links are present
    expect(screen.getByText('Curator')).toBeInTheDocument();
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Subscriptions')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();

    // Check that icons are present (by testing for links with icons)
    const homeLink = screen.getByRole('link', { name: /home/i });
    const subscriptionsLink = screen.getByRole('link', { name: /subscriptions/i });
    const settingsLink = screen.getByRole('link', { name: /settings/i });

    expect(homeLink).toBeInTheDocument();
    expect(subscriptionsLink).toBeInTheDocument();
    expect(settingsLink).toBeInTheDocument();
  });

  it('should have responsive classes to hide text labels on mobile', () => {
    render(
      <RootLayout>
        <div>Test content</div>
      </RootLayout>
    );

    // Test that labels have responsive classes
    const homeLabel = screen.getByText('Home');
    const subscriptionsLabel = screen.getByText('Subscriptions');
    const settingsLabel = screen.getByText('Settings');

    // Labels should have 'hidden sm:inline' classes to hide on small screens
    expect(homeLabel).toHaveClass('hidden', 'sm:inline');
    expect(subscriptionsLabel).toHaveClass('hidden', 'sm:inline');
    expect(settingsLabel).toHaveClass('hidden', 'sm:inline');
  });

  it('renders main content area', () => {
    render(
      <RootLayout>
        <div data-testid="test-content">Test content</div>
      </RootLayout>
    );

    expect(screen.getByTestId('test-content')).toBeInTheDocument();
  });

  it('renders footer with GitHub link', () => {
    render(
      <RootLayout>
        <div>Test content</div>
      </RootLayout>
    );

    // Check that the footer is present
    const githubLink = screen.getByRole('link', { name: /youtube curator v2/i });
    expect(githubLink).toBeInTheDocument();
    expect(githubLink).toHaveAttribute('href', 'https://github.com/bakkerme/youtube-curator-v2');
    expect(githubLink).toHaveAttribute('target', '_blank');
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer');
  });
});