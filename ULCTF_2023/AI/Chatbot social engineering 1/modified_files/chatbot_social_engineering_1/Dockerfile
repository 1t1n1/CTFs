# Use the official Python base image
FROM python:3.9-slim

# Set the working directory to /app
WORKDIR /app

# Copy the requirements file into the container
COPY requirements.txt .

# Install any required packages
RUN pip install --trusted-host pypi.python.org -r requirements.txt

# Copy the rest of the application code
COPY . .

# Expose the port your app runs on
EXPOSE 8000

# Use Gunicorn as the production server
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:8000", "app:app"]
