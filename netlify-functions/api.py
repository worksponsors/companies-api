import re
import json
import zlib
import requests
from functools import lru_cache

@lru_cache(maxsize=1)
def fetch_and_decompress_json():
    gz_url = 'https://data.worksponsors.co.uk/master.json.gz'
    response = requests.get(gz_url)
    compressed_data = response.content
    decompressed_data = zlib.decompress(compressed_data, zlib.MAX_WBITS | 16)
    json_data = json.loads(decompressed_data.decode('utf-8'))
    return json_data

def search_by_name(search_key, refined_data):
    regex_pattern = r"\b{}\b".format(re.escape(search_key))
    search_results = [item for item in refined_data if re.search(regex_pattern, item['name'], re.IGNORECASE)]
    return search_results

def get_companies(event, context):
    company_names_param = event['queryStringParameters'].get('companyNames')
    if not company_names_param:
        return {
            'statusCode': 400,
            'body': json.dumps({'error': 'Missing companyNames parameter'})
        }

    company_names = [name.strip() for name in company_names_param.split(',')]

    json_data = fetch_and_decompress_json()
    results = []
    for name in company_names:
        search_results = search_by_name(name, json_data)
        result = {
            'company': {
                'key': name,
                'count': len(search_results),
            }
        }
        if len(search_results) == 1:
            result['company']['exact_match'] = search_results[0]['name']
            result['company']['exact_rating'] = search_results[0].get('rating')
        results.append(result)

    return {
        'statusCode': 200,
        'body': json.dumps(results)
    }
