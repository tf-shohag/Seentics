'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle, Star, Zap, Crown } from 'lucide-react';
import Link from 'next/link';
import { useAuth } from '@/stores/useAuthStore';

const plans = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    description: "Perfect for getting started",
    icon: Star,
    popular: false,
    features: [
      "1 Website",
      "1,000 Monthly Events",
      "1 Workflow",
      "1 Funnel",
      "Basic Analytics",
      "Email Support",
      "Privacy Compliant"
    ],
    limitations: [
      "Limited integrations",
      "Basic reporting"
    ]
  },
  {
    name: "Standard",
    price: "$19",
    period: "per month",
    description: "Great for growing businesses",
    icon: Zap,
    popular: true,
    features: [
      "5 Websites",
      "100,000 Monthly Events",
      "10 Workflows",
      "10 Funnels",
      "Advanced Analytics",
      "Priority Support",
      "Custom Domains",
      "API Access",
      "Advanced Integrations",
      "A/B Testing"
    ],
    limitations: []
  },
  {
    name: "Pro",
    price: "$49",
    period: "per month",
    description: "For scaling teams and agencies",
    icon: Crown,
    popular: false,
    features: [
      "10 Websites",
      "500,000 Monthly Events",
      "30 Workflows",
      "30 Funnels",
      "Enterprise Analytics",
      "24/7 Priority Support",
      "White-label Options",
      "Advanced API Access",
      "Custom Integrations",
      "Advanced A/B Testing",
      "Team Collaboration",
      "Custom Reports"
    ],
    limitations: []
  }
];

export default function Pricing() {
  const { isAuthenticated, user } = useAuth();

  return (
    <section id="pricing" className="py-20 bg-white dark:bg-slate-950">
      <div className="container mx-auto px-4">
        
        {/* Header */}
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold mb-6 text-slate-900 dark:text-white">
            Simple, Transparent Pricing
          </h2>
          
          <p className="text-xl text-slate-600 dark:text-slate-400 max-w-3xl mx-auto">
            Start free and scale as you grow. No hidden fees, no surprises.
          </p>
        </div>
        
        {/* Pricing Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-7xl mx-auto mb-16">
          {plans.map((plan, index) => (
            <div key={index} className="relative">
              
              {/* Popular Badge */}
              {plan.popular && (
                <div className="absolute -top-12 left-1/2 transform -translate-x-1/2 z-10">
                  <div className="bg-blue-600 text-white px-4 py-2 rounded-full text-xs font-semibold">
                    Most Popular
                  </div>
                </div>
              )}
              
              {/* Card */}
              <Card className={`bg-white dark:bg-slate-800 shadow-lg hover:shadow-xl transition-all duration-300 border-0 flex flex-col h-full ${
                plan.popular ? 'ring-2 ring-blue-600 scale-105' : ''
              }`}>
                <CardHeader className="text-center pb-6">
                  
                  {/* Icon */}
                  <div className={`w-16 h-16 rounded-xl flex items-center justify-center mx-auto mb-6 ${
                    plan.popular 
                      ? 'bg-blue-600 text-white' 
                      : 'bg-slate-100 dark:bg-slate-700'
                  }`}>
                    <plan.icon className={`h-8 w-8 ${
                      plan.popular ? 'text-white' : 'text-slate-600 dark:text-slate-400'
                    }`} />
                  </div>
                  
                  {/* Plan Name */}
                  <CardTitle className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
                    {plan.name}
                  </CardTitle>
                  
                  {/* Price */}
                  <div className="mb-4">
                    <span className="text-4xl font-black text-slate-900 dark:text-white">
                      {plan.price}
                    </span>
                    <span className="text-slate-600 dark:text-slate-400 ml-2">
                      {plan.period}
                    </span>
                  </div>
                  
                  {/* Description */}
                  <p className="text-slate-600 dark:text-slate-400">
                    {plan.description}
                  </p>
                </CardHeader>
                
                <CardContent className="flex flex-col flex-grow">
                  {/* Features List */}
                  <ul className="space-y-3 mb-6 flex-grow">
                    {plan.features.map((feature, featureIndex) => (
                      <li key={featureIndex} className="flex items-center gap-3">
                        <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400 flex-shrink-0" />
                        <span className="text-slate-600 dark:text-slate-400 text-sm">
                          {feature}
                        </span>
                      </li>
                    ))}
                  </ul>
                  
                  {/* Limitations (if any) */}
                  {plan.limitations.length > 0 && (
                    <div className="pt-4 border-t border-slate-200 dark:border-slate-700 mb-6">
                      <p className="text-xs text-slate-500 dark:text-slate-500 mb-2">
                        Limitations:
                      </p>
                      <ul className="space-y-1">
                        {plan.limitations.map((limitation, limitIndex) => (
                          <li key={limitIndex} className="text-xs text-slate-500 dark:text-slate-500">
                            • {limitation}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                  
                  {/* CTA Button at bottom - always aligned */}
                  <div className="mt-auto">
                    {plan.name === 'Free' ? (
                      <Link href={isAuthenticated ? "/websites" : "/signup"}>
                        <Button 
                          className="w-full py-3 font-semibold bg-slate-900 dark:bg-white dark:text-slate-900 hover:bg-slate-800 dark:hover:bg-slate-100"
                        >
                          Start Free
                        </Button>
                      </Link>
                    ) : (
                      <a 
                        href={`https://seentics.lemonsqueezy.com/buy/39b59b36-94d3-40a5-821c-e31b6836345c${user?.id ? `?checkout[custom_data][user_id]=${user.id}&checkout[email]=${encodeURIComponent(user.email || '')}` : ''}`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Button 
                          className={`w-full py-3 font-semibold ${
                            plan.popular 
                              ? 'bg-blue-600 hover:bg-blue-700 text-white' 
                              : 'bg-slate-900 dark:bg-white dark:text-slate-900 hover:bg-slate-800 dark:hover:bg-slate-100'
                          }`}
                        >
                          Get {plan.name}
                        </Button>
                      </a>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          ))}
        </div>
        
        {/* Bottom Section */}
        <div className="text-center">
          <p className="text-slate-600 dark:text-slate-400 mb-4">
            All plans include our core features: real-time analytics, GDPR compliance, and no data limits.
          </p>
          
          <div className="flex items-center justify-center gap-8 text-sm text-slate-500 dark:text-slate-400">
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span>No setup fees</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span>Cancel anytime</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span>30-day money back</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
