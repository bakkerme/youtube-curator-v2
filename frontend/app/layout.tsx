import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "@/components/providers";
import Link from "next/link";
import { Bell, Home, Settings, FileText, Github } from "lucide-react";

export const metadata: Metadata = {
  title: "Curator",
  description: "Track your favorite YouTube channels",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased font-ui">
        <Providers>
          <div className="min-h-screen bg-background flex flex-col">
            {/* Navigation Header */}
            <header className="border-b">
              <div className="container mx-auto px-4">
                <nav className="flex items-center justify-between h-16">
                  <div className="flex items-center space-x-8">
                    <Link href="/" className="flex items-center space-x-2">
                      <div className="w-8 h-8 bg-red-600 rounded-full flex items-center justify-center">
                        <span className="text-white font-bold text-sm">YT</span>
                      </div>
                      <span className="font-semibold text-lg">Curator</span>
                    </Link>
                    
                    <div className="flex items-center space-x-6">
                      <Link href="/" className="flex items-center space-x-2 hover:text-gray-600 dark:hover:text-gray-300">
                        <Home className="w-4 h-4" />
                        <span className="hidden sm:inline">Home</span>
                      </Link>
                      <Link href="/subscriptions" className="flex items-center space-x-2 hover:text-gray-600 dark:hover:text-gray-300">
                        <Bell className="w-4 h-4" />
                        <span className="hidden sm:inline">Subscriptions</span>
                      </Link>
                      <Link href="/summary" className="flex items-center space-x-2 hover:text-gray-600 dark:hover:text-gray-300">
                        <FileText className="w-4 h-4" />
                        <span className="hidden sm:inline">Summary</span>
                      </Link>
                      <Link href="/notifications" className="flex items-center space-x-2 hover:text-gray-600 dark:hover:text-gray-300">
                        <Settings className="w-4 h-4" />
                        <span className="hidden sm:inline">Settings</span>
                      </Link>
                    </div>
                  </div>
                </nav>
              </div>
            </header>
            
            {/* Main Content */}
            <main className="container mx-auto px-4 py-8 flex-1">
              {children}
            </main>
            
            {/* Footer */}
            <footer className="border-t mt-auto">
              <div className="container mx-auto px-4 py-4">
                <div className="flex items-center justify-center">
                  <Link 
                    href="https://github.com/bakkerme/youtube-curator-v2" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="flex items-center space-x-2 text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
                  >
                    <Github className="w-4 h-4" />
                    <span className="text-sm">YouTube Curator v2</span>
                  </Link>
                </div>
              </div>
            </footer>
          </div>
        </Providers>
      </body>
    </html>
  );
}
