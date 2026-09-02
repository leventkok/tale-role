FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt serve.py ./
RUN pip install --no-cache-dir -r requirements.txt
# Install CUDA torch on the GPU host to match the card, e.g.
# pip install torch --index-url https://download.pytorch.org/whl/cu124
ENV RUNNER_HOST=0.0.0.0
EXPOSE 8091
CMD ["python", "serve.py", "--role", "storyteller"]
