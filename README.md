# Buenos Aires | GitOps Monitor

A modern, premium GitOps monitoring tool designed for clusters and infrastructure scripts. Built with Next.js 16, TailwindCSS v4, and featuring a state-of-the-art UI with full light and dark mode support.

![Buenos Aires Logo](buenosaires_logo.png)

## 🌟 Features

-   **Dual Theme Support**: Modern Light and Dark modes with a premium, glassmorphic aesthetic.
-   **Dynamic Dashboard**: Real-time overview of your environment with live stats (Total Projects, Active Tasks, Success Rate, Failed Jobs).
-   **Automated Project Sync**: Effortlessly add and monitor multiple Git repositories with automatic `.sh` script detection.
-   **Task Execution Engine**: Run infrastructure scripts directly from the UI with real-time status tracking and execution history.
-   **Mobile Responsive**: Fully optimized sidebar and layout for monitoring on the go.

## 🚀 Getting Started

### Local Installation

Prerequisites: Node.js v20+, Git, Bash.

1.  **Install dependencies**:
    ```bash
    npm install
    ```

2.  **Run the development server**:
    ```bash
    npm run dev
    ```

3.  **Open the app**:
    Navigate to [http://localhost:3000](http://localhost:3000).

### Docker Deployment

Buenos Aires is optimized for containerized environments.

#### Using Docker Compose (Recommended)

1.  **Run the application**:
    ```bash
    docker compose up -d
    ```

#### Using Docker CLI

1.  **Build the image**:
    ```bash
    docker build -t buenosaires .
    ```

2.  **Run the container**:
    ```bash
    docker run -p 3000:3000 \
      -v $(pwd)/repos:/app/repos \
      -v $(pwd)/db.json:/app/db.json \
      buenosaires
    ```

> [!CAUTION]
> **Common Error: Do not mount your source directory over `/app`** (e.g., `-v $(pwd):/app`). 
> The container contains a specialized production build that will be deleted/overwritten by your host source files if you do this, causing a `MODULE_NOT_FOUND` error.

## 🛠 Tech Stack

-   **Frontend**: Next.js 16 (App Router), React 19, TailwindCSS v4, Lucide React (Icons).
-   **Theming**: `next-themes` for seamless light/dark transitions.
-   **Backend**: Next.js API Routes (Server-side Git and Execution logic).
-   **Persistence**: Local JSON-based database (`db.json`) for simplicity and speed.
-   **Git Integration**: `simple-git` for repository synchronization.

## 📂 Project Structure

-   `/src/app`: Application routes and layout (Dashboard, Projects, Tasks, Settings).
-   `/src/components`: Reusable UI components (Sidebar, ThemeToggle, ThemeProvider).
-   `/src/lib`: Core services for Database, Git synchronization, and Task execution.
-   `/repos`: Local storage for cloned Git repositories.
