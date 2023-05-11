const axios = require('axios');
const pako = require('pako');
const cache = require('memory-cache');

exports.handler = async (event, context) => {
  try {
    const gzUrl = 'https://data.worksponsors.co.uk/master.json.gz';
    const cachedData = cache.get('jsonData');
    let jsonData;

    if (cachedData) {
      jsonData = cachedData;
    } else {
      const response = await axios.get(gzUrl, { responseType: 'arraybuffer' });
      const decompressedData = pako.inflate(response.data, { to: 'string' });
      jsonData = JSON.parse(decompressedData);
      cache.put('jsonData', jsonData);
    }

    const companyNamesParam = event.queryStringParameters.companyNames;
    if (!companyNamesParam) {
      return {
        statusCode: 400,
        body: JSON.stringify({ error: 'Missing companyNames parameter' }),
      };
    }

    const companyNames = companyNamesParam.split(',').map((name) => name.trim());

    const results = companyNames.map((name) => {
      const searchResults = searchByName(name, jsonData);
      return {
        company: {
          key: name,
          match: searchResults.length,
        },
      };
    });

    return {
      statusCode: 200,
      body: JSON.stringify(results),
    };
  } catch (error) {
    return {
      statusCode: 500,
      body: JSON.stringify({ error: 'Internal Server Error' }),
    };
  }
};

function searchByName(searchKey, refinedData) {
  const escapedSearchKey = searchKey.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const hashMap = {};

  refinedData.forEach((item) => {
    hashMap[item.name] = item;
  });

  const searchResults = Object.values(hashMap).filter((item) =>
    item.name.match(new RegExp(`\\b${escapedSearchKey}\\b`, 'i'))
  );

  return searchResults;
}
