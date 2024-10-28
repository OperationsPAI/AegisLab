# Prepare your running environment. This dockerfile needs to be as simple as possible, only use needed runtime dependencies. 
FROM python:3.10-slim AS runner
WORKDIR /app