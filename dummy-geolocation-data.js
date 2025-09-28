// Dummy Geolocation Data Generator
// This script sends batch events with geographic data to populate the analytics

const API_BASE = 'http://localhost:8080/api/v1/analytics';
const WEBSITE_ID = '68d8d06d2b9f57ed64fbbb6a';

// Sample countries with realistic visitor distributions
const countries = [
    { name: 'United States', weight: 35, cities: ['New York', 'Los Angeles', 'Chicago', 'Houston', 'Phoenix'] },
    { name: 'United Kingdom', weight: 15, cities: ['London', 'Manchester', 'Birmingham', 'Leeds', 'Glasgow'] },
    { name: 'Germany', weight: 12, cities: ['Berlin', 'Munich', 'Hamburg', 'Cologne', 'Frankfurt'] },
    { name: 'Canada', weight: 8, cities: ['Toronto', 'Vancouver', 'Montreal', 'Calgary', 'Ottawa'] },
    { name: 'France', weight: 7, cities: ['Paris', 'Lyon', 'Marseille', 'Toulouse', 'Nice'] },
    { name: 'Australia', weight: 6, cities: ['Sydney', 'Melbourne', 'Brisbane', 'Perth', 'Adelaide'] },
    { name: 'Japan', weight: 5, cities: ['Tokyo', 'Osaka', 'Kyoto', 'Yokohama', 'Nagoya'] },
    { name: 'Netherlands', weight: 4, cities: ['Amsterdam', 'Rotterdam', 'The Hague', 'Utrecht', 'Eindhoven'] },
    { name: 'India', weight: 3, cities: ['Mumbai', 'Delhi', 'Bangalore', 'Chennai', 'Kolkata'] },
    { name: 'Brazil', weight: 2, cities: ['São Paulo', 'Rio de Janeiro', 'Brasília', 'Salvador', 'Fortaleza'] },
    { name: 'Spain', weight: 2, cities: ['Madrid', 'Barcelona', 'Valencia', 'Seville', 'Bilbao'] },
    { name: 'Italy', weight: 1, cities: ['Rome', 'Milan', 'Naples', 'Turin', 'Florence'] }
];

// Sample pages
const pages = [
    '/',
    '/about',
    '/products',
    '/services',
    '/contact',
    '/blog',
    '/pricing',
    '/features',
    '/support',
    '/docs'
];

// Sample referrers
const referrers = [
    'https://google.com',
    'https://facebook.com',
    'https://twitter.com',
    'https://linkedin.com',
    'https://github.com',
    'Direct',
    'https://reddit.com',
    'https://stackoverflow.com'
];

// Sample browsers
const browsers = ['Chrome', 'Firefox', 'Safari', 'Edge', 'Opera'];
const devices = ['Desktop', 'Mobile', 'Tablet'];
const operatingSystems = ['Windows', 'macOS', 'Linux', 'iOS', 'Android'];

// Generate a random visitor ID
function generateVisitorId() {
    return 'visitor_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
}

// Generate a random session ID
function generateSessionId() {
    return 'session_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
}

// Get weighted random country
function getRandomCountry() {
    const totalWeight = countries.reduce((sum, country) => sum + country.weight, 0);
    let random = Math.random() * totalWeight;

    for (const country of countries) {
        random -= country.weight;
        if (random <= 0) {
            return country;
        }
    }
    return countries[0];
}

// Get random city from country
function getRandomCity(country) {
    return country.cities[Math.floor(Math.random() * country.cities.length)];
}

// Get random item from array
function getRandomItem(array) {
    return array[Math.floor(Math.random() * array.length)];
}

// Generate random timestamp within the last 7 days
function getRandomTimestamp() {
    const now = Date.now();
    const sevenDaysAgo = now - (7 * 24 * 60 * 60 * 1000);
    return new Date(sevenDaysAgo + Math.random() * (now - sevenDaysAgo));
}

// Generate a single event
function generateEvent() {
    const country = getRandomCountry();
    const city = getRandomCity(country);
    const visitorId = generateVisitorId();
    const sessionId = generateSessionId();
    const timestamp = getRandomTimestamp();

    return {
        website_id: WEBSITE_ID,
        visitor_id: visitorId,
        session_id: sessionId,
        page: getRandomItem(pages),
        event_type: 'pageview',
        referrer: getRandomItem(referrers),
        browser: getRandomItem(browsers),
        device: getRandomItem(devices),
        os: getRandomItem(operatingSystems),
        country: country.name,
        city: city,
        timestamp: timestamp.toISOString(),
        // Additional properties
        screen_width: Math.floor(Math.random() * 1000) + 1024,
        screen_height: Math.floor(Math.random() * 600) + 768,
        language: 'en-US',
        timezone: 'America/New_York',
        is_bot: false,
        is_new_user: Math.random() > 0.7,
        is_returning: Math.random() > 0.3
    };
}

// Generate batch of events
function generateEventBatch(count = 100) {
    const events = [];
    for (let i = 0; i < count; i++) {
        events.push(generateEvent());
    }
    return events;
}

// Send events to API
async function sendEvents(events) {
    try {
        console.log(`Sending ${events.length} events to ${API_BASE}/event/batch`);

        const response = await fetch(`${API_BASE}/event/batch`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': 'your-api-key-here' // Replace with actual API key if needed
            },
            body: JSON.stringify({ events })
        });

        if (response.ok) {
            const result = await response.json();
            console.log('✅ Events sent successfully:', result);
            return true;
        } else {
            const error = await response.text();
            console.error('❌ Failed to send events:', response.status, error);
            return false;
        }
    } catch (error) {
        console.error('❌ Error sending events:', error);
        return false;
    }
}

// Main function to generate and send dummy data
async function generateDummyGeolocationData() {
    console.log('🌍 Generating dummy geolocation data...');

    // Generate multiple batches to simulate realistic traffic
    const batches = [
        generateEventBatch(150), // High traffic batch
        generateEventBatch(100), // Medium traffic batch
        generateEventBatch(75),  // Lower traffic batch
        generateEventBatch(50),  // Small traffic batch
    ];

    let totalSent = 0;

    for (let i = 0; i < batches.length; i++) {
        console.log(`\n📦 Sending batch ${i + 1}/${batches.length}...`);
        const success = await sendEvents(batches[i]);

        if (success) {
            totalSent += batches[i].length;
            console.log(`✅ Batch ${i + 1} sent successfully`);
        } else {
            console.log(`❌ Batch ${i + 1} failed`);
        }

        // Wait a bit between batches to avoid overwhelming the server
        if (i < batches.length - 1) {
            await new Promise(resolve => setTimeout(resolve, 1000));
        }
    }

    console.log(`\n🎉 Dummy data generation complete!`);
    console.log(`📊 Total events sent: ${totalSent}`);
    console.log(`🌍 Countries represented: ${countries.length}`);
    console.log(`🏙️ Cities included: ${countries.reduce((sum, c) => sum + c.cities.length, 0)}`);
    console.log(`\n🔗 Test the API:`);
    console.log(`   ${API_BASE}/geolocation-breakdown/${WEBSITE_ID}?days=7`);
}

// Run the script
if (typeof window === 'undefined') {
  // Node.js environment - use built-in fetch (Node 18+) or https module
  if (typeof fetch === 'undefined') {
    // For older Node.js versions, use https module
    const https = require('https');
    const http = require('http');
    
    global.fetch = function(url, options = {}) {
      return new Promise((resolve, reject) => {
        const urlObj = new URL(url);
        const client = urlObj.protocol === 'https:' ? https : http;
        
        const req = client.request({
          hostname: urlObj.hostname,
          port: urlObj.port,
          path: urlObj.pathname + urlObj.search,
          method: options.method || 'GET',
          headers: options.headers || {}
        }, (res) => {
          let data = '';
          res.on('data', chunk => data += chunk);
          res.on('end', () => {
            resolve({
              ok: res.statusCode >= 200 && res.statusCode < 300,
              status: res.statusCode,
              json: () => Promise.resolve(JSON.parse(data)),
              text: () => Promise.resolve(data)
            });
          });
        });
        
        req.on('error', reject);
        
        if (options.body) {
          req.write(options.body);
        }
        
        req.end();
      });
    };
  }
  
  generateDummyGeolocationData();
} else {
  // Browser environment
  console.log('Run this in Node.js or copy the functions to browser console');
}

// Export for browser use
if (typeof window !== 'undefined') {
    window.generateDummyGeolocationData = generateDummyGeolocationData;
    window.generateEventBatch = generateEventBatch;
    window.sendEvents = sendEvents;
}
