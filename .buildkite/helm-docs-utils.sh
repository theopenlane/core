#!/bin/bash
# Shared utilities for Helm documentation generation

# Function to install task if not available
install_task() {
  if ! command -v task >/dev/null 2>&1; then
    echo "Installing task..."
    local task_version="3.40.1"
    local download_url="https://github.com/go-task/task/releases/download/v${task_version}/task_linux_amd64.tar.gz"

    wget -q "$download_url" -O /tmp/task.tar.gz
    tar -xzf /tmp/task.tar.gz -C /tmp
    mv /tmp/task /usr/local/bin/task
    chmod +x /usr/local/bin/task
    rm -f /tmp/task.tar.gz
    echo "✅ task v${task_version} installed"
  else
    echo "✅ task already available"
  fi
}

# Function to install helm-docs if not available
install_helm_docs() {
  if ! command -v helm-docs >/dev/null 2>&1; then
    echo "Installing helm-docs..."
    local helm_docs_version="1.15.0"
    local download_url="https://github.com/norwoodj/helm-docs/releases/download/v${helm_docs_version}/helm-docs_${helm_docs_version}_Linux_x86_64.tar.gz"

    wget -q "$download_url" -O /tmp/helm-docs.tar.gz
    tar -xzf /tmp/helm-docs.tar.gz -C /tmp
    mv /tmp/helm-docs /usr/local/bin/helm-docs
    chmod +x /usr/local/bin/helm-docs
    rm -f /tmp/helm-docs.tar.gz
    echo "✅ helm-docs v${helm_docs_version} installed"
  else
    echo "✅ helm-docs already available"
  fi
}

# Function to generate documentation and add to git
generate_docs_and_commit() {
  echo "📚 Generating documentation..."

  # Install task if needed
  install_task

  # Install helm-docs if needed
  install_helm_docs

  # Generate documentation
  if task docs; then
    # Add any generated documentation files
    git add .
    echo "✅ Documentation generated and added to commit"
    return 0
  else
    echo "⚠️  Documentation generation failed, continuing without docs"
    return 1
  fi
}