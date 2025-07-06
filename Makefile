.PHONY: download
download:
	@echo "Downloading sql-formatter.min.js from unpkg.com"
	@curl -L "https://unpkg.com/sql-formatter@latest/dist/sql-formatter.min.js" -o assets/sql-formatter.min.js