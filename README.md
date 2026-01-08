# 🦁 TeachMe AI

**TeachMe AI** is a safe, locally-hosted AI tutor and study buddy designed specifically for children. It acts as a friendly wrapper around powerful Large Language Models (LLMs), ensuring a controlled and educational environment for kids to ask questions, learn new concepts, and explore their curiosity without the risks associated with unmoderated internet access.

![TeachMe AI Preview](assets/teachme-preview.png)

## 🚀 Purpose

In an age where AI is becoming ubiquitous, it is crucial to provide children with tools that are both educational and safe. Expensive subscriptions or cloud-based services often come with privacy concerns or lack specific safety guardrails for young minds. **TeachMe AI** solves this by:

* **Running Locally**: All data stays on your machine. No conversations are sent to the cloud.
* **Safety First**: The system prompt is engineered to be a "kind, patient, and safe AI tutor," explicitly instructed to refuse harmful or age-inappropriate topics.
* **Cost-Effective**: Uses open-source models (via Ollama) and free software, running on standard consumer hardware.
* **Simple UI**: A "kid-first" design with large text, playful colors, and an easy-to-use chat interface.

## ✨ Features

* **Safe Chat Interface**: A friendly, moderated chat experience powered by the `phi3` model (customizable).
* **Conversation History**: Automatically saves chats so kids can revisit previous lessons or fun conversations.
* **Organization**: Rename conversations to keep track of different topics (e.g., "Math Homework," "Dinosaur Facts").
* **Privacy**: Zero external data tracking. Everything is stored in a local MongoDB database.
* **Full Control**: Parents can easily delete conversations or reset the history.

## 🛠️ Technology Stack

This project is built with a modern, performance-oriented stack:

* **Backend**: [Go (Golang) 1.23](https://go.dev/) - Fast, strongly typed, and efficient.
* **Frontend**: [React](https://react.dev/) + [Vite](https://vitejs.dev/) - Responsive and snappy user interface.
* **Database**: [MongoDB](https://www.mongodb.com/) - flexible storage for chat logs and sessions.
* **AI Engine**: [Ollama](https://ollama.com/) - The easiest way to run LLMs locally.

## 🏁 Getting Started

### Prerequisites

* [Docker](https://www.docker.com/) & Docker Compose
* [Ollama](https://ollama.com/) installed on your host machine (or accessible via network).

### Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/darkraiden/TeachMe.git
    cd TeachMe
    ```

2. **Pull the AI Model:**
    Ensure Ollama is running and pull the `phi3` model (or your preferred small/safe model).

    ```bash
    ollama pull phi3
    ```

3. **Start the Application:**
    Run the entire stack using Docker Compose:

    ```bash
    docker compose up --build
    ```

    Or simply using the Makefile:

    ```bash
    make docker-up
    ```

4. **Access the App:**
    Open your browser and navigate to `http://localhost:3000`.

### Usage

* **Ask a Question**: Type in the box and hit send!
* **New Chat**: Click `+ New Chat` to start a fresh topic.
* **Rename**: Hover over a session in the sidebar and click the **pencil icon** ✎ to give it a custom name.
* **Delete**: Hover over a session and click the **trash icon** 🗑️ to remove it.

## 🧪 Development & Testing

The project includes a `Makefile` to simplify common tasks.

* **Run All Tests**:

    ```bash
    make test
    ```

* **Run Backend Tests**:

    ```bash
    make test-backend
    ```

* **Run Frontend Tests**:

    ```bash
    make test-frontend
    ```

## 🤝 Contributing

Contributions are welcome! Whether it's adding a new feature, improving the safety prompts, or fixing a bug, feel free to open a Pull Request.

## 📄 License

[MIT License](LICENSE)
