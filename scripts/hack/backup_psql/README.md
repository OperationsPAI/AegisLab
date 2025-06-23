
crontab -e

```bash
#!/bin/bash
PROJECT_DIR="/home/nn/workspace/rcabench"         # 修改为你的项目路径
LOG_FILE="$PROJECT_DIR/temp/pg_backup.log"
SCRIPT="$PROJECT_DIR/scripts/hack/backup_psql/cli.py"
CRON_CMD="0 * * * * $SCRIPT pg-backup >> $LOG_FILE 2>&1"
mkdir -p "$(dirname "$LOG_FILE")"
(crontab -l 2>/dev/null | grep -v -F "$SCRIPT"; echo "$CRON_CMD") | crontab -
echo "Cron job added:"
echo "$CRON_CMD"
```