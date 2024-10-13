import pandas as pd
from drain3 import TemplateMiner
import os
import sys
import re
from drain3.redis_persistence import RedisPersistence
from drain3.file_persistence import FilePersistence
import multiprocessing
import logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

logger = logging.getLogger(__name__)
template_miner = TemplateMiner(
        persistence_handler=RedisPersistence(
            redis_host="localhost", redis_port=6379, redis_db=1, redis_key="pass", redis_pass=None, is_ssl=False
        )
    )


template_miner.load_state()
print('\n'.join([c.get_template() for c in template_miner.drain.clusters]),file=open('templates.txt','w'))
