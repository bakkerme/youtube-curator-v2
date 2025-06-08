import React from 'react'
import { render, screen } from '@/lib/test-utils'
import VideoCard from './VideoCard'
import { VideoEntry, Channel } from '@/lib/types'
import { formatDistanceToNow } from 'date-fns'

// Mock next/image since it has specific requirements in tests
jest.mock('next/image', () => ({
  __esModule: true,
  default: (props: { src: string; alt: string; width?: number; height?: number; className?: string; fill?: boolean }) => {
    // Filter out Next.js specific props that don't belong on img elements
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { fill: _fill, ...imgProps } = props
    // eslint-disable-next-line @next/next/no-img-element
    return <img {...imgProps} alt={props.alt || ''} />
  },
}))

describe('VideoCard', () => {
  const mockChannels: Channel[] = [
    {
      id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
      title: 'Google Developers',
    },
  ]

  const mockVideo: VideoEntry = {
    channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
    cachedAt: new Date().toISOString(),
    entry: {
      title: 'Introduction to React Testing',
      link: {
        Href: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
        Rel: 'alternate',
      },
      id: 'yt:video:dQw4w9WgXcQ',
      published: new Date().toISOString(),
      content: 'A comprehensive introduction to testing React applications',
      author: {
        name: 'Google Developers',
        uri: 'https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw',
      },
      mediaGroup: {
        mediaThumbnail: {
          URL: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg',
          Width: '1280',
          Height: '720',
        },
        mediaTitle: 'Introduction to React Testing',
        mediaContent: {
          URL: 'https://www.youtube.com/v/dQw4w9WgXcQ',
          Type: 'application/x-shockwave-flash',
          Width: '640',
          Height: '390',
        },
        mediaDescription: 'A comprehensive introduction to testing React applications',
      },
    },
  }

  it('renders video information correctly', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    // Check title is rendered
    expect(screen.getByText('Introduction to React Testing')).toBeInTheDocument()
    
    // Check channel name is rendered
    expect(screen.getByText('Google Developers')).toBeInTheDocument()
    
    // Check time ago is rendered (will be "less than a minute ago" for new date)
    expect(screen.getByText(/ago/)).toBeInTheDocument()
    
    // Check Watch button exists
    expect(screen.getByRole('link', { name: /watch/i })).toBeInTheDocument()
  })

  it('opens video link in new tab', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    const watchLink = screen.getByRole('link', { name: /watch/i })
    expect(watchLink).toHaveAttribute('href', 'https://www.youtube.com/watch?v=dQw4w9WgXcQ')
    expect(watchLink).toHaveAttribute('target', '_blank')
    expect(watchLink).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('renders thumbnail image', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    const thumbnail = screen.getByAltText('Introduction to React Testing')
    expect(thumbnail).toHaveAttribute('src', 'https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg')
  })

  it('handles missing thumbnail gracefully', () => {
    const videoWithoutThumbnail: VideoEntry = {
      ...mockVideo,
      entry: {
        ...mockVideo.entry,
        mediaGroup: {
          mediaThumbnail: {
            URL: '',
            Width: '0',
            Height: '0',
          },
          mediaTitle: mockVideo.entry.title,
          mediaContent: {
            URL: '',
            Type: '',
            Width: '0',
            Height: '0',
          },
          mediaDescription: '',
        },
      },
    }
    
    render(<VideoCard video={videoWithoutThumbnail} channels={mockChannels} />)
    
    const thumbnail = screen.getByAltText('Introduction to React Testing')
    expect(thumbnail).toHaveAttribute('src', '/placeholder-video.svg')
  })

  it('handles unknown channel', () => {
    const videoWithUnknownChannel = {
      ...mockVideo,
      channelId: 'unknown-channel-id',
    }
    
    render(<VideoCard video={videoWithUnknownChannel} channels={mockChannels} />)
    
    expect(screen.getByText('Unknown Channel')).toBeInTheDocument()
  })

  it('handles missing video title', () => {
    const videoWithoutTitle = {
      ...mockVideo,
      entry: {
        ...mockVideo.entry,
        title: '',
      },
    }
    
    render(<VideoCard video={videoWithoutTitle} channels={mockChannels} />)
    
    expect(screen.getByText('Untitled Video')).toBeInTheDocument()
  })

  it('formats published date correctly', () => {
    const specificDate = new Date('2024-01-15T10:00:00Z')
    const videoWithSpecificDate = {
      ...mockVideo,
      entry: {
        ...mockVideo.entry,
        published: specificDate.toISOString(),
      },
    }
    
    render(<VideoCard video={videoWithSpecificDate} channels={mockChannels} />)
    
    // Calculate expected time ago
    const expectedTimeAgo = formatDistanceToNow(specificDate, { addSuffix: true })
    expect(screen.getByText(expectedTimeAgo)).toBeInTheDocument()
  })

  it('applies correct CSS classes for dark mode', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    // Check for dark mode classes on the card
    const card = screen.getByText('Introduction to React Testing').closest('div[class*="bg-white"]')
    expect(card).toHaveClass('dark:bg-gray-800')
    
    // Check text has dark mode classes - the parent div has the dark mode class
    const channelText = screen.getByText('Google Developers')
    const textContainer = channelText.closest('div[class*="text-gray-600"]')
    expect(textContainer).toHaveClass('dark:text-gray-400')
  })

  it('has hover effect on card', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    const card = screen.getByText('Introduction to React Testing').closest('div[class*="bg-white"]')
    expect(card).toHaveClass('hover:shadow-lg')
    expect(card).toHaveClass('transition-all')
  })

  it('has correct styling for watch button', () => {
    render(<VideoCard video={mockVideo} channels={mockChannels} />)
    
    const watchButton = screen.getByRole('link', { name: /watch/i })
    expect(watchButton).toHaveClass('bg-red-600')
    expect(watchButton).toHaveClass('hover:bg-red-700')
    expect(watchButton).toHaveClass('text-white')
  })
}) 