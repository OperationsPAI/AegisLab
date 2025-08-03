# MySQL Backup Tool

A comprehensive command-line tool for MySQL database backup and restore operations with support for compression, transaction consistency, and various MySQL features.

## Features

- 🚀 **Easy Installation**: Automatic MySQL client tools installation
- 💾 **Smart Backup**: Support for compression, single transactions, routines, and triggers
- 🔄 **Flexible Restore**: Restore from latest backup or specific files
- 🛡️ **Safety Features**: Database existence checks and confirmation prompts
- 📊 **Rich Output**: Beautiful console output with progress indicators
- ⚙️ **Configurable**: Extensive command-line options and Makefile shortcuts

## Quick Start

### 1. Installation

```bash
# Install MySQL client tools
uv run python cli.py install-tools

# Or use Makefile
make install
```

### 2. Basic Usage

```bash
# Create a backup
uv run python cli.py backup

# List available backups
uv run python cli.py list

# Restore latest backup
uv run python cli.py restore

# Or use Makefile shortcuts
make backup
make restore-local
```

## Commands

### Core Commands

|  **Command**   | **Description**  |
|  ----  | ----  |
| `install-tools` | Install MySQL client tools automatically |
| `check-version` | Check MySQL server version and connectivity |
| `backup` | Create database backup using mysqldump |
| `restore` | Restore database from backup file  |
| `list` | List all available backup files |
| `help` | Show comprehensive help information |

### Backup Options

```bash
uv run python cli.py backup [OPTIONS]
```

|  **Option**   | **Default**  |  **Description**  |
|  ----  | ----  | ----  |
| `--host` | `10.10.10.220` | MySQL server host |
| `--port` | `32206` | MySQL server port |
| `--user` | `root` | MySQL username |
| `--password` | `yourpassword` | MySQL password |
| `--database` | `rcabench` | Database name |
| `--compress` | `True` | 	Enable network compression |
| `--single-transaction` | `True` | Use single transaction (InnoDB) |
| `--routines` | `True` | Include stored procedures/functions |
| `--triggers` | `True` | Include triggers |

### Restore Options

```bash
uv run python cli.py restore [OPTIONS]
```

|  **Option**   | **Default**  |  **Description**  |
|  ----  | ----  | ----  |
| `--backup-file` | `None` | Specific backup file to restore (optional) |
| `--host` | `10.10.10.220` | MySQL server host |
| `--port` | `32206` | MySQL server port |
| `--user` | `root` | MySQL username |
| `--password` | `yourpassword` | MySQL password |
| `--database` | `rcabench` | Database name |
| `--force` | `True` | 	Force overwrite without confirmation |

## Usage Examples

### Basic Operations

```bash
# Check server connectivity
uv run python cli.py check-version

# Create full backup with all features
uv run python cli.py backup --compress --single-transaction --routines --triggers

# Quick backup without routines/triggers
uv run python cli.py backup --compress --single-transaction

# Force restore without confirmation
uv run python cli.py restore --force

# Restore specific backup file
uv run python cli.py restore --backup-file backup_20231215_143022.sql.gz
```

### Cross-Server Migration

```bash
# 1. Backup from remote server
uv run python cli.py backup \
    --host 10.10.10.220 \
    --port 32206 \
    --user root \
    --password yourpassword \
    --database rcabench

# 2. Restore to local server
uv run python cli.py restore \
    --host 127.0.0.1 \
    --port 3306 \
    --user root \
    --password yourpassword \
    --database rcabench \
    --force
```

### Using Makefile Shortcuts
```bash
# Show all available commands
make help

# Quick operations
make install          # Install tools
make check-remote     # Check remote server
make backup           # Create backup
make restore-local    # Restore to local
make list             # List backups
make clean            # Clean old backups

# Advanced operations
make migrate          # Full migration (backup → restore)
```

## Configuration

### Default Configuration

The tool uses these default values (configurable in [cli.py](./cli.py)):

```python
DEFAULT_MYSQL_HOST = "10.10.10.220"
DEFAULT_MYSQL_PORT = "32206"
DEFAULT_MYSQL_USER = "root"
DEFAULT_MYSQL_PASSWORD = "yourpassword"
DEFAULT_MYSQL_DB = "rcabench"
```

### Makefile Configuration

Modify variables in [Makefile](./Makefile):

```makefile
# Remote Database (Source)
REMOTE_HOST := 10.10.10.220
REMOTE_PORT := 32206
REMOTE_USER := root
REMOTE_PASSWORD := yourpassword
REMOTE_DB := rcabench

# Local Database (Target)
LOCAL_HOST := 127.0.0.1
LOCAL_PORT := 3306
LOCAL_USER := root
LOCAL_PASSWORD := yourpassword
LOCAL_DB := rcabench
```

## File Structure

```bash
backup_mysql/
├── cli.py             # Main CLI application
├── Makefile           # Make shortcuts and automation
├── README.md          # This documentation
└── temp/
    └── backup_mysql/  # Backup files storage
        ├── rcabench_mysql_backup_20231215_143022.sql.gz
        └── rcabench_mysql_backup_20231215_150130.sql.gz
```

## Backup File Format

Backup files are named with timestamp for easy identification:

```plain
{database}_mysql_backup_{YYYYMMDD_HHMMSS}.sql.gz
```

Example: `rcabench_mysql_backup_20231215_143022.sql.gz`

## Requirements

### System Requirements

- Python 3.13+
- MySQL client tools (`mysql`, `mysqldump`)
- `gzip` for compression
- `uv` for Python environment management


### Python Dependencies
- `typer>=0.16.0` - CLI framework
- `rich` - Rich text and beautiful formatting

### Installation Methods

#### Ubuntu/Debian

```bash
# Option 1: Using the tool
uv run python cli.py install-tools

# Option 2: Manual installation
sudo apt update
sudo apt install mysql-client
```

### macOS

```bash
# Option 1: Using the tool
uv run python cli.py install-tools

# Option 2: Manual installation
brew install mysql
```

## Best Practices

### Backup Strategy

1. **Regular Backups**: Schedule daily backups using cron
2. **Compression**: Always use --compress for large databases
3. **Consistency**: Use --single-transaction for InnoDB tables
4. **Complete Backups**: Include --routines and --triggers for full schema

### Restore Safety

1. **Test First**: Always test restores in a non-production environment
2. **Backup Before Restore**: Create a backup before major restores
3. **Use --force Carefully**: Only use --force when you're certain
4. **Verify Data**: Check data integrity after restore

### Storage Management

```bash
# Clean old backups (7+ days)
make clean

# Check disk usage
make status

# Manual cleanup
find ./temp/backup_mysql -name "*_mysql_backup_*" -mtime +7 -delete
```

## Troubleshooting

Common Issues

### Connection Problems

```bash
# Test connectivity
uv run python cli.py check-version

# Check if MySQL client is installed
which mysql mysqldump

# Install missing tools
make install
```

### Permission Issues

```sql
# Ensure user has proper privileges
GRANT SELECT, LOCK TABLES, SHOW VIEW ON database.* TO 'user'@'host';
GRANT ALL PRIVILEGES ON database.* TO 'user'@'host';  # For restore
```

### Large Database Backups

```bash
# Use compression and optimize settings
uv run python cli.py backup \
  --compress \
  --single-transaction \
  --no-routines \
  --no-triggers
```

## Advanced Usage

### Automated Backups

Create a cron job for automated backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /path/to/backup_mysql && make backup
```

### Custom Scripts

Example integration script:

```bash
#!/bin/bash
# backup_and_notify.sh

cd /path/to/backup_mysql

# Create backup
if make backup; then
    echo "Backup successful: $(date)" | mail -s "DB Backup OK" admin@company.com
else
    echo "Backup failed: $(date)" | mail -s "DB Backup FAILED" admin@company.com
fi
```

## Support
For support and questions:

1. Check this README
2. Run `uv run python cli.py help`
3. Use `make help` for Makefile commands
4. Review the troubleshooting section
5. Create an issue in the repository

---

**Happy Backing Up!** 🚀