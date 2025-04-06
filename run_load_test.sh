#!/bin/bash

# Create test users
uv run create_users.py

# Run stress test (Ctrl+C to stop)
uv run stress_test.py

# Cleanup after test
uv run cleanup.py
rm test_users.txt