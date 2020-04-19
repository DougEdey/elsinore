#!/bin/bash
echo "Setting up databases..."
diesel setup --database-url="test-elsinore.db"
diesel setup --database-url="elsinore.db"

echo "Done"