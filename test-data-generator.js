#!/usr/bin/env node

/**
 * Seentics Analytics Test Data Generator
 * 
 * This script generates realistic test data for your analytics system.
 * It sends various types of events including pageviews, custom events, and UTM campaigns.
 */

const https = require('https');
const http = require('http');

// Configuration
const CONFIG = {
    API_HOST: 'http://localhost:8080',
    SITE_ID: '68dcee11f2bb1aed5ecf502b', // Your website ID
    BATCH_SIZE: 10,
    DELAY_BETWEEN_BATCHES: 1000, // 1 second
    TOTAL_EVENTS: 500, // Increased for better charts
    DAYS_BACK: 30, // Generate data for last 30 days
    EVENTS_PER_DAY_MIN: 10,
    EVENTS_PER_DAY_MAX: 50
};

// Sample data pools
const SAMPLE_DATA = {
    pages: [
        '/68dcee11f2bb1aed5ecf502b/analytics',
        '/68dcee11f2bb1aed5ecf502b/dashboard',
        '/68dcee11f2bb1aed5ecf502b/settings',
        '/68dcee11f2bb1aed5ecf502b/funnels',
        '/68dcee11f2bb1aed5ecf502b/workflows',
        '/68dcee11f2bb1aed5ecf502b/billing',
        '/68dcee11f2bb1aed5ecf502b/support',
        '/68dcee11f2bb1aed5ecf502b/privacy'
    ],

    referrers: [
        'https://google.com/search',
        'https://facebook.com',
        'https://twitter.com',
        'https://linkedin.com',
        'https://github.com',
        'https://stackoverflow.com',
        '',  // Direct traffic
        'https://reddit.com'
    ],

    browsers: ['Chrome', 'Firefox', 'Safari', 'Edge'],
    devices: ['Desktop', 'Mobile', 'Tablet'],
    os: ['Windows', 'macOS', 'Linux', 'iOS', 'Android'],

    countries: [
        'United States', 'United Kingdom', 'Canada', 'Germany', 'France',
        'Japan', 'Australia', 'Brazil', 'India', 'Netherlands'
    ],

    cities: [
        'New York', 'London', 'Toronto', 'Berlin', 'Paris',
        'Tokyo', 'Sydney', 'São Paulo', 'Mumbai', 'Amsterdam'
    ],

    utmSources: [
        'google', 'facebook', 'twitter', 'linkedin', 'email',
        'newsletter', 'github', 'reddit', 'direct', null
    ],

    utmMediums: [
        'organic', 'cpc', 'social', 'email', 'referral',
        'display', 'affiliate', null
    ],

    utmCampaigns: [
        'spring_sale', 'product_launch', 'brand_awareness', 'retargeting',
        'newsletter_signup', 'free_trial', 'black_friday', null
    ],

    customEvents: [
        'button_click', 'form_submit', 'video_play', 'download',
        'signup', 'purchase', 'search', 'share', 'like', 'comment'
    ]
};

// Utility functions
function randomChoice(array) {
    return array[Math.floor(Math.random() * array.length)];
}

function randomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

function generateVisitorId() {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

function generateSessionId() {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

// Generate realistic date distribution for historical data
function generateHistoricalDates() {
    const dates = [];
    const now = new Date();

    for (let i = 0; i < CONFIG.DAYS_BACK; i++) {
        const date = new Date(now);
        date.setDate(date.getDate() - i);

        // Create realistic traffic patterns
        const dayOfWeek = date.getDay(); // 0 = Sunday, 6 = Saturday
        const isWeekend = dayOfWeek === 0 || dayOfWeek === 6;
        const isMonday = dayOfWeek === 1;

        // Traffic patterns: Lower on weekends, spike on Mondays, steady during week
        let trafficMultiplier = 1.0;
        if (isWeekend) {
            trafficMultiplier = 0.6; // 40% less traffic on weekends
        } else if (isMonday) {
            trafficMultiplier = 1.4; // 40% more traffic on Mondays
        } else {
            trafficMultiplier = 1.0 + (Math.random() * 0.4 - 0.2); // ±20% variation
        }

        // Add some seasonal trends (more recent = slightly more traffic)
        const recencyBoost = 1 + (i / CONFIG.DAYS_BACK) * 0.3;
        trafficMultiplier *= recencyBoost;

        const baseEvents = randomInt(CONFIG.EVENTS_PER_DAY_MIN, CONFIG.EVENTS_PER_DAY_MAX);
        const eventsForDay = Math.round(baseEvents * trafficMultiplier);

        dates.push({
            date: date,
            events: eventsForDay,
            isWeekend: isWeekend
        });
    }

    return dates.reverse(); // Oldest first
}

// Generate random time within a specific day
function generateRandomTimeInDay(baseDate) {
    const date = new Date(baseDate);

    // Create realistic hourly distribution (more traffic during business hours)
    const hour = Math.random() < 0.7
        ? randomInt(9, 17) // 70% during business hours (9 AM - 5 PM)
        : randomInt(0, 23); // 30% other times

    const minute = randomInt(0, 59);
    const second = randomInt(0, 59);

    date.setHours(hour, minute, second, randomInt(0, 999));
    return date;
}

// Generate realistic event data
function generatePageviewEvent(visitorId, sessionId, timestamp = new Date()) {
    const timeOnPage = randomInt(5, 300); // 5 seconds to 5 minutes

    const event = {
        website_id: CONFIG.SITE_ID,
        visitor_id: visitorId,
        session_id: sessionId,
        event_type: 'pageview',
        page: randomChoice(SAMPLE_DATA.pages),
        time_on_page: timeOnPage,
        timestamp: timestamp.toISOString(),
        browser: randomChoice(SAMPLE_DATA.browsers),
        device: randomChoice(SAMPLE_DATA.devices),
        os: randomChoice(SAMPLE_DATA.os)
    };

    // Add referrer (70% chance)
    if (Math.random() > 0.3) {
        event.referrer = randomChoice(SAMPLE_DATA.referrers);
    }

    // Add UTM parameters (40% chance)
    if (Math.random() > 0.6) {
        const utmSource = randomChoice(SAMPLE_DATA.utmSources);
        const utmMedium = randomChoice(SAMPLE_DATA.utmMediums);
        const utmCampaign = randomChoice(SAMPLE_DATA.utmCampaigns);

        if (utmSource) event.utm_source = utmSource;
        if (utmMedium) event.utm_medium = utmMedium;
        if (utmCampaign) event.utm_campaign = utmCampaign;
    }

    return event;
}

function generateCustomEvent(visitorId, sessionId, timestamp = new Date()) {
    const eventType = randomChoice(SAMPLE_DATA.customEvents);

    const event = {
        website_id: CONFIG.SITE_ID,
        visitor_id: visitorId,
        session_id: sessionId,
        event_type: eventType,
        page: randomChoice(SAMPLE_DATA.pages),
        timestamp: timestamp.toISOString(),
        browser: randomChoice(SAMPLE_DATA.browsers),
        device: randomChoice(SAMPLE_DATA.devices),
        os: randomChoice(SAMPLE_DATA.os),
        properties: generateEventProperties(eventType)
    };

    return event;
}

// Generate realistic properties based on event type
function generateEventProperties(eventType) {
    const baseProperties = {
        user_agent: randomChoice(['Chrome/91.0', 'Firefox/89.0', 'Safari/14.1', 'Edge/91.0']),
        screen_resolution: randomChoice(['1920x1080', '1366x768', '1440x900', '1536x864']),
        timestamp: new Date().toISOString()
    };

    switch (eventType) {
        case 'button_click':
            return {
                ...baseProperties,
                button_text: randomChoice(['Get Started', 'Learn More', 'Sign Up', 'Download', 'Contact Us', 'Try Free']),
                button_id: randomChoice(['cta-primary', 'hero-button', 'nav-signup', 'footer-contact']),
                button_color: randomChoice(['blue', 'green', 'red', 'orange']),
                position: randomChoice(['header', 'hero', 'sidebar', 'footer']),
                click_count: randomInt(1, 5)
            };

        case 'form_submit':
            return {
                ...baseProperties,
                form_name: randomChoice(['contact_form', 'newsletter_signup', 'login_form', 'registration_form']),
                form_fields: randomInt(3, 8),
                validation_errors: randomInt(0, 2),
                time_to_complete: randomInt(30, 300), // seconds
                form_type: randomChoice(['lead_generation', 'authentication', 'feedback', 'subscription'])
            };

        case 'video_play':
            return {
                ...baseProperties,
                video_title: randomChoice(['Product Demo', 'Tutorial Video', 'Company Overview', 'Customer Testimonial']),
                video_duration: randomInt(60, 600), // seconds
                video_quality: randomChoice(['720p', '1080p', '480p']),
                autoplay: randomChoice([true, false]),
                video_id: `video_${randomInt(1000, 9999)}`
            };

        case 'download':
            return {
                ...baseProperties,
                file_name: randomChoice(['whitepaper.pdf', 'product_guide.pdf', 'pricing.xlsx', 'demo.zip']),
                file_size: randomInt(100, 5000), // KB
                file_type: randomChoice(['pdf', 'xlsx', 'zip', 'docx']),
                download_source: randomChoice(['header', 'content', 'sidebar', 'popup'])
            };

        case 'signup':
            return {
                ...baseProperties,
                signup_method: randomChoice(['email', 'google', 'facebook', 'github']),
                plan_type: randomChoice(['free', 'basic', 'premium', 'enterprise']),
                referral_code: Math.random() > 0.7 ? `REF${randomInt(1000, 9999)}` : null,
                newsletter_opt_in: randomChoice([true, false]),
                terms_accepted: true
            };

        case 'purchase':
            return {
                ...baseProperties,
                product_id: `prod_${randomInt(1000, 9999)}`,
                product_name: randomChoice(['Basic Plan', 'Pro Plan', 'Enterprise Plan', 'Add-on Package']),
                amount: randomInt(9, 299),
                currency: 'USD',
                payment_method: randomChoice(['credit_card', 'paypal', 'stripe', 'bank_transfer']),
                discount_code: Math.random() > 0.8 ? `SAVE${randomInt(10, 50)}` : null
            };

        default:
            return {
                ...baseProperties,
                event_category: randomChoice(['engagement', 'conversion', 'navigation', 'interaction']),
                event_value: randomInt(1, 100),
                custom_data: {
                    feature_flag: randomChoice(['A', 'B', 'control']),
                    user_segment: randomChoice(['new', 'returning', 'premium']),
                    experiment_id: `exp_${randomInt(100, 999)}`
                }
            };
    }
}

// Send batch of events to API
async function sendEventBatch(events) {
    const payload = {
        siteId: CONFIG.SITE_ID,
        domain: 'localhost',
        events: events
    };

    const data = JSON.stringify(payload);

    return new Promise((resolve, reject) => {
        const url = new URL(`${CONFIG.API_HOST}/api/v1/analytics/event/batch`);
        const options = {
            hostname: url.hostname,
            port: url.port,
            path: url.pathname,
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Content-Length': Buffer.byteLength(data)
            }
        };

        const client = url.protocol === 'https:' ? https : http;

        const req = client.request(options, (res) => {
            let responseData = '';

            res.on('data', (chunk) => {
                responseData += chunk;
            });

            res.on('end', () => {
                if (res.statusCode >= 200 && res.statusCode < 300) {
                    console.log(`✅ Batch sent successfully (${events.length} events) - Status: ${res.statusCode}`);
                    resolve({ statusCode: res.statusCode, data: responseData });
                } else {
                    console.error(`❌ Batch failed - Status: ${res.statusCode}, Response: ${responseData}`);
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

// Generate events for a specific day
function generateEventsForDay(dayData) {
    const events = [];
    const { date, events: eventCount } = dayData;

    // Generate multiple sessions per day (1-5 sessions)
    const sessionsPerDay = randomInt(1, Math.min(5, Math.ceil(eventCount / 3)));

    for (let session = 0; session < sessionsPerDay; session++) {
        const visitorId = generateVisitorId();
        const sessionId = generateSessionId();
        const eventsInSession = Math.ceil(eventCount / sessionsPerDay);

        for (let i = 0; i < eventsInSession; i++) {
            const eventTime = generateRandomTimeInDay(date);

            // 80% pageviews, 20% custom events
            if (Math.random() > 0.2) {
                events.push(generatePageviewEvent(visitorId, sessionId, eventTime));
            } else {
                events.push(generateCustomEvent(visitorId, sessionId, eventTime));
            }
        }
    }

    return events;
}

// Generate a batch of mixed events (for backward compatibility)
function generateEventBatch(size = CONFIG.BATCH_SIZE) {
    const events = [];
    const visitorId = generateVisitorId();
    const sessionId = generateSessionId();

    for (let i = 0; i < size; i++) {
        // 80% pageviews, 20% custom events
        if (Math.random() > 0.2) {
            events.push(generatePageviewEvent(visitorId, sessionId));
        } else {
            events.push(generateCustomEvent(visitorId, sessionId));
        }
    }

    return events;
}

// Main execution function
async function generateTestData() {
    console.log('🚀 Starting Seentics Analytics Test Data Generator');
    console.log(`📊 Configuration:`);
    console.log(`   - API Host: ${CONFIG.API_HOST}`);
    console.log(`   - Site ID: ${CONFIG.SITE_ID}`);
    console.log(`   - Days Back: ${CONFIG.DAYS_BACK}`);
    console.log(`   - Events Per Day: ${CONFIG.EVENTS_PER_DAY_MIN}-${CONFIG.EVENTS_PER_DAY_MAX}`);
    console.log(`   - Batch Size: ${CONFIG.BATCH_SIZE}`);
    console.log(`   - Delay: ${CONFIG.DELAY_BETWEEN_BATCHES}ms`);
    console.log('');

    // Generate historical date distribution
    console.log('📅 Generating historical date distribution...');
    const historicalDates = generateHistoricalDates();
    const totalExpectedEvents = historicalDates.reduce((sum, day) => sum + day.events, 0);

    console.log(`📈 Generated ${historicalDates.length} days with ${totalExpectedEvents} total events`);
    console.log('');

    let totalSent = 0;
    let successfulBatches = 0;
    let failedBatches = 0;
    let dayCount = 0;

    // Process each day
    for (const dayData of historicalDates) {
        dayCount++;
        const dateStr = dayData.date.toISOString().split('T')[0];
        const isWeekend = dayData.isWeekend ? '🏖️' : '💼';

        console.log(`${isWeekend} Day ${dayCount}/${historicalDates.length}: ${dateStr} (${dayData.events} events)`);

        // Generate events for this day
        const dayEvents = generateEventsForDay(dayData);

        // Send events in batches
        const dayBatches = Math.ceil(dayEvents.length / CONFIG.BATCH_SIZE);

        for (let batchNum = 0; batchNum < dayBatches; batchNum++) {
            const startIdx = batchNum * CONFIG.BATCH_SIZE;
            const endIdx = Math.min(startIdx + CONFIG.BATCH_SIZE, dayEvents.length);
            const batchEvents = dayEvents.slice(startIdx, endIdx);

            try {
                await sendEventBatch(batchEvents);
                totalSent += batchEvents.length;
                successfulBatches++;

                // Small delay between batches
                if (batchNum < dayBatches - 1) {
                    await new Promise(resolve => setTimeout(resolve, CONFIG.DELAY_BETWEEN_BATCHES / 2));
                }

            } catch (error) {
                console.error(`❌ Batch failed for ${dateStr}:`, error.message);
                failedBatches++;
            }
        }

        console.log(`   ✅ Sent ${dayEvents.length} events for ${dateStr}`);

        // Delay between days (except for the last one)
        if (dayCount < historicalDates.length) {
            await new Promise(resolve => setTimeout(resolve, CONFIG.DELAY_BETWEEN_BATCHES));
        }
    }

    console.log('');
    console.log('🎉 Historical test data generation completed!');
    console.log(`📊 Summary:`);
    console.log(`   - Days processed: ${dayCount}`);
    console.log(`   - Total events sent: ${totalSent}/${totalExpectedEvents}`);
    console.log(`   - Successful batches: ${successfulBatches}`);
    console.log(`   - Failed batches: ${failedBatches}`);
    console.log(`   - Success rate: ${((successfulBatches / (successfulBatches + failedBatches)) * 100).toFixed(1)}%`);
    console.log('');
    console.log('📈 Your analytics charts should now show beautiful historical trends!');
}

// CLI interface
if (require.main === module) {
    // Parse command line arguments
    const args = process.argv.slice(2);

    // Override config from command line
    args.forEach(arg => {
        if (arg.startsWith('--days=')) {
            CONFIG.DAYS_BACK = parseInt(arg.split('=')[1]);
        } else if (arg.startsWith('--min-events=')) {
            CONFIG.EVENTS_PER_DAY_MIN = parseInt(arg.split('=')[1]);
        } else if (arg.startsWith('--max-events=')) {
            CONFIG.EVENTS_PER_DAY_MAX = parseInt(arg.split('=')[1]);
        } else if (arg.startsWith('--batch-size=')) {
            CONFIG.BATCH_SIZE = parseInt(arg.split('=')[1]);
        } else if (arg.startsWith('--delay=')) {
            CONFIG.DELAY_BETWEEN_BATCHES = parseInt(arg.split('=')[1]);
        } else if (arg.startsWith('--site-id=')) {
            CONFIG.SITE_ID = arg.split('=')[1];
        } else if (arg.startsWith('--host=')) {
            CONFIG.API_HOST = arg.split('=')[1];
        } else if (arg === '--help' || arg === '-h') {
            console.log('Seentics Analytics Historical Test Data Generator');
            console.log('');
            console.log('Usage: node test-data-generator.js [options]');
            console.log('');
            console.log('Options:');
            console.log('  --days=N           Days of historical data to generate (default: 30)');
            console.log('  --min-events=N     Minimum events per day (default: 10)');
            console.log('  --max-events=N     Maximum events per day (default: 50)');
            console.log('  --batch-size=N     Events per batch (default: 10)');
            console.log('  --delay=N          Delay between batches in ms (default: 1000)');
            console.log('  --site-id=ID       Website ID to use (default: current)');
            console.log('  --host=URL         API host URL (default: http://localhost:8080)');
            console.log('  --help, -h         Show this help message');
            console.log('');
            console.log('Examples:');
            console.log('  node test-data-generator.js --days=7 --min-events=20 --max-events=100');
            console.log('  node test-data-generator.js --site-id=your-site-id --days=14');
            console.log('  node test-data-generator.js --days=60 --batch-size=20');
            console.log('');
            console.log('Features:');
            console.log('  • Realistic traffic patterns (weekends vs weekdays)');
            console.log('  • Business hours distribution (9 AM - 5 PM peak)');
            console.log('  • UTM campaign data (~40% of events)');
            console.log('  • Mixed pageviews (80%) and custom events (20%)');
            console.log('  • Multiple sessions per day with realistic visitor IDs');
            process.exit(0);
        }
    });

    // Start generation
    generateTestData().catch(error => {
        console.error('💥 Fatal error:', error.message);
        process.exit(1);
    });
}
