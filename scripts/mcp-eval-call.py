#!/usr/bin/env python3
"""File bridge to the SDK client hosted by TestPostgresStdio, for model evals."""
import json
import os
from pathlib import Path
import sys
import time
import uuid

root = Path(os.environ.get('MCP_EVAL_DIR', '/private/tmp/mcp-model-eval'))
request_id = uuid.uuid4().hex
payload = json.loads(sys.argv[1])
payload['request_id'] = request_id
request = root / (request_id + '.request.json')
temporary = request.with_suffix('.tmp')
temporary.write_text(json.dumps(payload))
temporary.replace(request)
response = root / (request_id + '.response.json')
for _ in range(200):
    if response.exists():
        print(response.read_text())
        break
    time.sleep(0.05)
else:
    sys.exit('MCP eval host did not respond within 10 seconds')
