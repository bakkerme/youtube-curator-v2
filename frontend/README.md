# YouTube Curator v2 Frontend

This is the Next.js frontend for YouTube Curator v2, providing a web interface to manage YouTube channel subscriptions.

## Features

- **Channel Management**: Add and remove YouTube channels to monitor
- **Dark Mode Support**: Automatic theme switching based on system preferences
- **Responsive Design**: Works on desktop and mobile devices
- **Real-time Updates**: Uses React Query for efficient data fetching and caching

## Development

### Prerequisites

- Node.js 20 or higher
- npm or yarn
- Backend API running on port 8080

### Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Create a `.env.local` file:
   ```bash
   NEXT_PUBLIC_API_URL=http://localhost:8080/api
   ```

3. Run the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

## Production Build

### Using Docker

The frontend is configured to run in Docker as part of the docker-compose setup:

```bash
# From the project root
docker-compose up
```

The frontend will be available at http://localhost:3000

### Standalone Build

To build for production:

```bash
npm run build
npm start
```

## Configuration

The frontend requires the following environment variable:

- `NEXT_PUBLIC_API_URL`: The URL of the backend API (default: `http://localhost:8080/api`)

When running in Docker, this is automatically configured to communicate with the backend container.

## Project Structure

```
frontend/
├── app/                    # Next.js app directory
│   ├── layout.tsx         # Root layout with navigation
│   ├── page.tsx           # Home page
│   ├── subscriptions/     # Subscription management page
│   └── notifications/     # Notification settings (placeholder)
├── components/            # Reusable React components
├── lib/                   # Utilities and API client
│   ├── api.ts            # API client with axios
│   └── types.ts          # TypeScript type definitions
├── public/               # Static assets
└── design/               # UI design mockups
```

## API Integration

The frontend communicates with the backend API using the following endpoints:

- `GET /api/channels` - List all subscribed channels
- `POST /api/channels` - Add a new channel
- `DELETE /api/channels/{id}` - Remove a channel
- `GET /api/config/interval` - Get check interval
- `PUT /api/config/interval` - Set check interval

## Technologies Used

- **Next.js 15**: React framework with App Router
- **TypeScript**: Type-safe development
- **Tailwind CSS**: Utility-first CSS framework
- **React Query**: Data fetching and state management
- **Axios**: HTTP client
- **Lucide React**: Icon library

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
