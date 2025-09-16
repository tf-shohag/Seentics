import axios from 'axios';

// Create Lemon Squeezy API instance
export const lemonSqueezyApiInstance = axios.create({
  baseURL: 'https://api.lemonsqueezy.com/v1',
  headers: {
    'Accept': 'application/vnd.api+json',
    'Content-Type': 'application/vnd.api+json',
    'Authorization': `Bearer ${process.env.LEMONSQUEEZY_API_KEY}`
  }
});

// Verify webhook signature
export const verifyWebhookSignature = (payload, signature, secret) => {
  const crypto = require('crypto');
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload, 'utf8');
  const digest = hmac.digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature, 'hex'),
    Buffer.from(digest, 'hex')
  );
};

// Get customer from Lemon Squeezy
export const getCustomer = async (customerId) => {
  try {
    const response = await lemonSqueezyApiInstance.get(`/customers/${customerId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching customer:', error);
    throw error;
  }
};

// Get subscription from Lemon Squeezy
export const getSubscription = async (subscriptionId) => {
  try {
    const response = await lemonSqueezyApiInstance.get(`/subscriptions/${subscriptionId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching subscription:', error);
    throw error;
  }
};

// Create customer in Lemon Squeezy
export const createCustomer = async (customerData) => {
  try {
    const response = await lemonSqueezyApiInstance.post('/customers', {
      data: {
        type: 'customers',
        attributes: customerData
      }
    });
    return response.data;
  } catch (error) {
    console.error('Error creating customer:', error);
    throw error;
  }
};
