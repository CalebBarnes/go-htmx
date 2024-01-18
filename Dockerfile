FROM golang:1.21

# Set destination for COPY
WORKDIR /app

# Copy the entire contents of the current directory
COPY ./ ./

# Install dependencies and build the application
RUN apt-get update && \
    apt-get install -y nodejs npm && \
    npm i -g pnpm && \
    pnpm install && \
    go mod tidy && \
    pnpm run build

EXPOSE 42069

CMD ["pnpm", "run", "start"]