import Link from 'next/link';
import { Bell } from 'lucide-react';

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <div className="w-24 h-24 bg-red-600 rounded-full flex items-center justify-center mb-8">
        <span className="text-white font-bold text-4xl">YT</span>
      </div>
      
      <h1 className="text-4xl font-bold mb-4">Welcome to YT Curator</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8 max-w-2xl">
        Your self-hosted YouTube channel monitoring service. Track your favorite channels and get notified about new uploads via email.
      </p>
      
      <div className="flex gap-4">
        <Link
          href="/subscriptions"
          className="px-6 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700 
                   transition-colors flex items-center gap-2"
        >
          <Bell className="w-5 h-5" />
          Manage Subscriptions
        </Link>
      </div>
      
      <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-8 max-w-4xl text-left">
        <div className="p-6">
          <h3 className="font-semibold mb-2">🔔 Real-time Monitoring</h3>
          <p className="text-gray-600 dark:text-gray-400">
            Automatically check your subscribed channels for new videos at your preferred interval.
          </p>
        </div>
        <div className="p-6">
          <h3 className="font-semibold mb-2">📧 Email Notifications</h3>
          <p className="text-gray-600 dark:text-gray-400">
            Get notified via email whenever your favorite channels upload new content.
          </p>
        </div>
        <div className="p-6">
          <h3 className="font-semibold mb-2">🔐 Self-Hosted</h3>
          <p className="text-gray-600 dark:text-gray-400">
            Complete control over your data with a fully self-hosted solution.
          </p>
        </div>
      </div>
    </div>
  );
}
