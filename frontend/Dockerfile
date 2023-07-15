# Use an official Nginx image as the base
FROM nginx:alpine

# Remove default Nginx configuration
RUN rm -rf /etc/nginx/conf.d

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the built React app from the local machine to the container
COPY build /usr/share/nginx/html

# Expose a port for the container
EXPOSE 80

# Start the Nginx web server
CMD ["nginx", "-g", "daemon off;"]