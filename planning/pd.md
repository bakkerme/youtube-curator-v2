# Youtube Curator v2 Product Document

## 1. Overview

Youtube Curator v2 is a platform designed to help users stay updated on new video releases from their favorite YouTube channels.

## 2. Goals

The primary goals for Youtube Curator v2 are to:

*   **Empower Individual Users:** Provide a tool for individuals to take control of their YouTube content consumption, free from the distractions of the native platform.
*   **Promote Digital Wellbeing:** Support users in curating their media diet by delivering updates for desired channels at predictable times.
*   **Ensure Self-Hostability and Control:** Make the platform easy to set up, deploy, and update in a self-hosted environment, giving users full ownership and control over their data and notifications.
*   **Provide a User-Friendly Interface:** Include a basic UI to simplify the process of adding and managing channels.
*   **Offer Configuration Flexibility:** Allow users to easily configure the interval at which the system checks for new videos.

## 3. Key Features

The platform will include the following core functionalities:

*   **Channel Subscription:** Allow users to provide a list of YouTube channels they want to monitor.
*   **Scheduled Checks:** Periodically search for new videos on the subscribed channels.
*   **Email Notifications:** Compile new video information into an email.
*   **Configurable Delivery:** Send the email to a specified address using provided SMTP details.
*   **Basic Web UI:** Provide a simple web interface for:
    *   Adding and removing YouTube channel IDs/URLs.
    *   Viewing the list of currently monitored channels.
    *   Configuring the video check interval.
    *   Inputting and saving SMTP server details and the recipient email address.

## 4. Technical Considerations (High Level)

Based on the project goals and desired self-hostability, the following technical approach is planned:

*   **Data Source:** Initially leverage YouTube's RSS feeds for fetching new video information, building upon existing code and experience.
*   **Database:** Utilize BadgerDB for storing application data such as channel lists, user configurations, and tracking the last checked video.
*   **Scheduling:** Implement the periodic checking mechanism using Go's built-in timer system. A fallback to external scheduling like cron is an option if needed.
*   **Technology Stack:** The backend will be developed using Go with the Echo web framework. The frontend will be built with Next.js and styled using Tailwind CSS.
*   **Deployment:** The primary deployment method will focus on providing a Docker container to ensure ease of self-hosting and updates.

## 5. Open Questions & Future Work

**Open Questions:**

*   **RSS Feed Limitations:** How will we handle potential limitations or changes to YouTube's RSS feed functionality, including rate limits or data availability?
*   **Error Handling & Monitoring:** How will the system handle errors during feed fetching, email sending, or other operations? What level of logging and monitoring will be necessary for a self-hosted environment?
*   **Configuration Management:** How will the application manage and store sensitive configurations like SMTP credentials securely, especially in a Dockerized environment?
*   **Initial Data Sync:** When a channel is added, how far back should the system check for existing videos? Should it only notify for videos published *after* the channel is added?

**Potential Future Features:**

*  **Watch Page:** Embed video into web page
*   **YouTube Data API Integration:** Explore migrating from or supplementing RSS with the official YouTube Data API for potentially richer data, better reliability, and rate limit management.
*   **Advanced Scheduling Options:** Allow more granular control over the checking schedule beyond a single interval (e.g., specific times of day, different intervals for different channels).
*   **Email Content Customization:** Provide options for users to customize the format and content of the notification emails.
*   **Import/Export Channel List:** Allow users to easily import and export their list of subscribed channels.
*   **UI Enhancements:** Develop more advanced UI features like searching, filtering, or grouping channels.