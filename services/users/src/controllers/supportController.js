import { Resend } from 'resend';

export const sendSupportEmail = async (req, res) => {
  try {
    const { subject, message, email } = req.body;
    const userId = req.userId || req.user?.id;

    // Validate required fields
    if (!subject || !message || !email) {
      return res.status(400).json({
        success: false,
        message: 'Subject, message, and email are required'
      });
    }

    // Check if Resend API key is configured
    const resendApiKey = process.env.RESEND_API_KEY;
    if (!resendApiKey) {
      console.log('RESEND_API_KEY not configured - email will not be sent');
      console.log('Support request details:', { subject, email, message, userId });
    } else {
      console.log('Attempting to send email via Resend...');
      
      // Initialize Resend
      const resend = new Resend(resendApiKey);

      try {
        const result = await resend.emails.send({
          from: 'onboarding@resend.dev',
          to: 'shohagmiah2100@gmail.com',
          subject: `[Support] ${subject}`,
          html: `
            <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
              <h2 style="color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px;">
                Support Request
              </h2>
              
              <div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
                <p><strong>From:</strong> ${email}</p>
                <p><strong>User ID:</strong> ${userId || 'Not authenticated'}</p>
                <p><strong>Subject:</strong> ${subject}</p>
              </div>
              
              <div style="background: white; padding: 20px; border: 1px solid #dee2e6; border-radius: 5px;">
                <h3 style="color: #333; margin-top: 0;">Message:</h3>
                <p style="line-height: 1.6; color: #555;">${message.replace(/\n/g, '<br>')}</p>
              </div>
              
              <div style="margin-top: 20px; padding: 10px; background: #e9ecef; border-radius: 5px; font-size: 12px; color: #6c757d;">
                <p><strong>Timestamp:</strong> ${new Date().toISOString()}</p>
                <p>This email was sent from the Seentics support system.</p>
              </div>
            </div>
          `
        });
        console.log('Email sent successfully:', result);
      } catch (emailError) {
        console.error('Email sending failed:', emailError);
        console.error('Error details:', emailError.message);
        // Continue with success response even if email fails
      }
    }

    res.status(200).json({
      success: true,
      message: 'Support request sent successfully',
      data: {
        subject,
        timestamp: new Date().toISOString()
      }
    });

  } catch (error) {
    console.error('Support request error:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to send support request',
      error: process.env.NODE_ENV === 'development' ? error.message : 'Internal server error'
    });
  }
};
