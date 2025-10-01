#!/usr/bin/env node

const http = require('http');

// Test events with different dates to trigger partition creation
const testEvents = [
    {
        siteId: '68dce91da04a92762aaab5f5',
        domain: 'localhost',
        events: [{
            website_id: '68dce91da04a92762aaab5f5',
            visitor_id: 'test_visitor_sep',
            session_id: 'test_session_sep',
            event_type: 'pageview',
            page: '/test-september',
            timestamp: '2025-09-15T10:00:00.000Z', // September 2025
            browser: 'Chrome',
            device: 'Desktop',
            os: 'Windows'
        }]
    },
    {
        siteId: '68dce91da04a92762aaab5f5',
        domain: 'localhost',
        events: [{
            website_id: '68dce91da04a92762aaab5f5',
            visitor_id: 'test_visitor_oct',
            session_id: 'test_session_oct',
            event_type: 'pageview',
            page: '/test-october',
            timestamp: '2025-10-01T10:00:00.000Z', // October 2025
            browser: 'Chrome',
            device: 'Desktop',
            os: 'Windows'
        }]
    },
    {
        siteId: '68dce91da04a92762aaab5f5',
        domain: 'localhost',
        events: [{
            website_id: '68dce91da04a92762aaab5f5',
            visitor_id: 'test_visitor_nov',
            session_id: 'test_session_nov',
            event_type: 'pageview',
            page: '/test-november',
            timestamp: '2025-11-01T10:00:00.000Z', // November 2025
            browser: 'Chrome',
            device: 'Desktop',
            os: 'Windows'
        }]
    }
];

async function sendEvent(eventData) {
    const data = JSON.stringify(eventData);
    
    const options = {
        hostname: 'localhost',
        port: 8080,
        path: '/api/v1/analytics/event/batch',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(data),
            'x-api-key': 'your-global-api-key-for-service-communication'
        }
    };

    return new Promise((resolve, reject) => {
        const req = http.request(options, (res) => {
            let responseData = '';
            
            res.on('data', (chunk) => {
                responseData += chunk;
            });
            
            res.on('end', () => {
                const timestamp = eventData.events[0].timestamp;
                const month = new Date(timestamp).toISOString().substring(0, 7);
                
                if (res.statusCode >= 200 && res.statusCode < 300) {
                    console.log(`✅ Event for ${month} sent successfully - Status: ${res.statusCode}`);
                    resolve({ statusCode: res.statusCode, data: responseData });
                } else {
                    console.log(`❌ Event for ${month} failed - Status: ${res.statusCode}, Response: ${responseData}`);
                    reject(new Error(`HTTP ${res.statusCode}: ${responseData}`));
                }
            });
        });

        req.on('error', (err) => {
            console.error('❌ Request error:', err.message);
            reject(err);
        });

        req.write(data);
        req.end();
    });
}

async function testPartitionCreation() {
    console.log('🧪 Testing automatic partition creation...');
    console.log('📡 Sending events for different months to trigger partition creation\n');
    
    for (let i = 0; i < testEvents.length; i++) {
        const event = testEvents[i];
        const month = new Date(event.events[0].timestamp).toISOString().substring(0, 7);
        
        console.log(`📅 Testing partition for ${month}...`);
        
        try {
            await sendEvent(event);
            
            // Small delay between requests
            if (i < testEvents.length - 1) {
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
        } catch (error) {
            console.error(`❌ Failed to send event for ${month}:`, error.message);
        }
    }
    
    console.log('\n🎉 Partition creation test completed!');
    console.log('📊 Check the analytics service logs to see if partitions were created automatically.');
}

testPartitionCreation();
