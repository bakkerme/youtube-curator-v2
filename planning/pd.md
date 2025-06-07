# YouTube Curator v2 Product Document

## 1. Overview

YouTube Curator v2 is a comprehensive platform designed to help users stay updated on new video releases from their favorite YouTube channels while providing intelligent insights and interactive viewing experiences.

## 2. Goals

The primary goals for YouTube Curator v2 are to:

*   **Empower Individual Users:** Provide a tool for individuals to take control of their YouTube content consumption, free from the distractions of the native platform.
*   **Promote Digital Wellbeing:** Support users in curating their media diet by delivering updates for desired channels at predictable times.
*   **Ensure Self-Hostability and Control:** Make the platform easy to set up, deploy, and update in a self-hosted environment, giving users full ownership and control over their data and notifications.
*   **Provide a User-Friendly Interface:** Include a comprehensive UI to simplify the process of adding and managing channels, viewing videos, and accessing insights.
*   **Offer Configuration Flexibility:** Allow users to easily configure the interval at which the system checks for new videos.
*   **Enable Intelligent Content Analysis:** Leverage AI to provide meaningful insights about video content and community engagement.
*   **Foster Interactive Learning:** Allow users to engage with video content through AI-powered conversations and analysis.

## 3. Key Features

The platform will include the following core functionalities:

### Core Features
*   **Channel Subscription:** Allow users to provide a list of YouTube channels they want to monitor.
*   **Scheduled Checks:** Periodically search for new videos on the subscribed channels.
*   **Email Notifications:** Compile new video information into an email.
*   **Configurable Delivery:** Send the email to a specified address using provided SMTP details.

### Web Interface Features
*   **Channel Management UI:** 
    *   Adding and removing YouTube channel IDs/URLs.
    *   Viewing the list of currently monitored channels.
    *   Configuring the video check interval.
    *   Inputting and saving SMTP server details and the recipient email address.
*   **Latest Videos UI:**
    *   Grid/list view of all fetched videos from subscribed channels.
    *   Search functionality to filter videos by title or channel.
    *   "Today" filter for recent content.
    *   Pagination for navigating through video collections.

### Advanced Playback Features
*   **Embedded Video Player:**
    *   Clean, distraction-free YouTube video embedding.
    *   Full-screen and standard viewing modes.
    *   Seamless integration with the platform's UI.
*   **LLM-Driven Comment Analysis:**
    *   Intelligent analysis of YouTube comments using Large Language Models.
    *   Sentiment analysis and key theme extraction.
    *   Summary of community reactions and discussions.
    *   Identification of common questions and concerns.
*   **Interactive Video Chat:**
    *   AI-powered chat interface for discussing video content.
    *   Context-aware conversations about the current video.
    *   Ability to ask questions about video topics, creators, or related content.
    *   Educational support for complex topics covered in videos.

## 4. Technical Considerations (High Level)

Based on the project goals and desired self-hostability, the following technical approach is planned:

*   **Data Source:** Initially leverage YouTube's RSS feeds for fetching new video information, building upon existing code and experience.
*   **Database:** Utilize BadgerDB for storing application data such as channel lists, user configurations, video metadata, and tracking the last checked video.
*   **Scheduling:** Implement the periodic checking mechanism using Go's built-in timer system. A fallback to external scheduling like cron is an option if needed.
*   **Technology Stack:** 
    *   Backend: Go with the Echo web framework
    *   Frontend: Next.js and styled using Tailwind CSS
    *   AI Integration: OpenAI API or compatible LLM service for comment analysis and chat functionality
*   **Deployment:** The primary deployment method will focus on providing a Docker container to ensure ease of self-hosting and updates.
*   **YouTube Integration:** YouTube Data API integration for enhanced video metadata and comment retrieval.

## 5. User Experience Flow

### Enhanced Video Discovery and Viewing
1. **Discovery:** Users receive email notifications or browse the Latest Videos UI to find new content.
2. **Selection:** Users click on a video to enter the enhanced playback experience.
3. **Viewing:** Users watch the embedded video in a clean, distraction-free environment.
4. **Analysis:** Users access AI-driven comment analysis to understand community sentiment and key discussion points.
5. **Interaction:** Users engage with the AI chat to discuss video content, ask questions, or explore related topics.
6. **Learning:** Users gain deeper insights through interactive conversations about the video's subject matter.
7. **Summarisation:** Videos can be summarised by AI, assisting users to understand content and decide if it's relevant.

## 6. Open Questions & Future Work

**Open Questions:**

*   **RSS Feed Limitations:** How will we handle potential limitations or changes to YouTube's RSS feed functionality, including rate limits or data availability?
*   **Error Handling & Monitoring:** How will the system handle errors during feed fetching, email sending, or other operations? What level of logging and monitoring will be necessary for a self-hosted environment?
*   **Configuration Management:** How will the application manage and store sensitive configurations like SMTP credentials and LLM API keys securely, especially in a Dockerized environment?
*   **Initial Data Sync:** When a channel is added, how far back should the system check for existing videos? Should it only notify for videos published *after* the channel is added?
*   **LLM Integration:** Which LLM provider should be the default? How will we handle API rate limits and costs for self-hosted users?
*   **Comment Analysis Scope:** How many comments should be analyzed per video? Should we focus on top comments, recent comments, or a representative sample?

**Potential Future Features:**

*   **Advanced AI Features:**
    *   Cross-video topic tracking and recommendations
    *   Automated tagging and categorization
    *   Creator analysis and content pattern recognition
*   **Enhanced Playback:**
    *   Playlist creation and management
    *   Watch history and progress tracking
    *   Video bookmarking and notes
    *   Offline viewing capabilities
*   **Social Features:**
    *   Share insights and discussions
    *   Export analysis reports
    *   Collaborative viewing and discussion rooms
*   **Integration Enhancements:**
    *   Custom prompt templates for analysis
    *   Advanced scheduling options
    *   Import/export channel lists
    *   Enhanced email content customization