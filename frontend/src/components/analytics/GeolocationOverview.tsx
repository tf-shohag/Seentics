'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatNumber } from '@/lib/analytics-api';
import { Globe, MapPin } from 'lucide-react';
import dynamic from 'next/dynamic';
import Image from 'next/image';
import { useState } from 'react';

// Dynamically import WorldMap to avoid SSR issues
const WorldMap = dynamic(() => import('./WorldMap'), {
    ssr: false,
    loading: () => (
        <div className="h-[32rem] rounded-lg flex items-center justify-center">
            <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                <p className="text-sm text-gray-500">Loading map...</p>
            </div>
        </div>
    )
});

interface TopItem {
    name: string;
    count: number;
    percentage: number;
}

interface GeolocationData {
    countries: TopItem[];
    continents: TopItem[];
    regions: TopItem[];
    cities: TopItem[];
}

interface GeolocationOverviewProps {
    data?: GeolocationData;
    isLoading?: boolean;
    className?: string;
}

export function GeolocationOverview({ data, isLoading = false, className = '' }: GeolocationOverviewProps) {
    const [geoTab, setGeoTab] = useState<string>('map');

    // Helper function to get country flag
    const getCountryFlag = (countryName: string): string => {
        const countryMap: Record<string, string> = {
            'United States': 'US',
            'United States of America': 'US',
            'USA': 'US',
            'US': 'US',
            'Bangladesh': 'BD',
            'India': 'IN',
            'China': 'CN',
            'United Kingdom': 'GB',
            'UK': 'GB',
            'Germany': 'DE',
            'France': 'FR',
            'Canada': 'CA',
            'Australia': 'AU',
            'Japan': 'JP',
            'Brazil': 'BR',
            'Russia': 'RU',
            'South Korea': 'KR',
            'Italy': 'IT',
            'Spain': 'ES',
            'Netherlands': 'NL',
            'Sweden': 'SE',
            'Norway': 'NO',
            'Denmark': 'DK',
            'Finland': 'FI',
            'Switzerland': 'CH',
            'Austria': 'AT',
            'Belgium': 'BE',
            'Poland': 'PL',
            'Mexico': 'MX',
            'Argentina': 'AR',
            'Chile': 'CL',
            'Colombia': 'CO',
            'Peru': 'PE',
            'Venezuela': 'VE',
            'Thailand': 'TH',
            'Vietnam': 'VN',
            'Indonesia': 'ID',
            'Philippines': 'PH',
            'Malaysia': 'MY',
            'Singapore': 'SG',
            'Pakistan': 'PK',
            'Turkey': 'TR',
            'Egypt': 'EG',
            'South Africa': 'ZA',
            'Nigeria': 'NG',
            'Kenya': 'KE',
            'Morocco': 'MA',
            'Algeria': 'DZ',
            'Tunisia': 'TN',
            'Ghana': 'GH',
            'Ethiopia': 'ET',
            'Uganda': 'UG',
            'Tanzania': 'TZ',
            'Zimbabwe': 'ZW',
            'Botswana': 'BW',
            'Namibia': 'NA'
        };

        const countryCode = countryMap[countryName] || countryName.substring(0, 2).toUpperCase();
        return `https://flagcdn.com/w40/${countryCode.toLowerCase()}.png`;
    };

    // Helper function to get continent emoji
    const getContinentEmoji = (continent: string): string => {
        const continentMap: Record<string, string> = {
            'North America': '🌎',
            'South America': '🌎',
            'Europe': '🌍',
            'Asia': '🌏',
            'Africa': '🌍',
            'Australia': '🌏',
            'Oceania': '🌏',
            'Antarctica': '🧊'
        };
        return continentMap[continent] || '🌍';
    };

    if (isLoading) {
        return (
            <Card className={`bg-card border-0 shadow-md mb-6 rounded-none ${className}`}>
                <CardHeader>
                    <div className="animate-pulse">
                        <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-48 mb-2"></div>
                        <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-64"></div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="animate-pulse space-y-4">
                        {[...Array(5)].map((_, i) => (
                            <div key={i} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded"></div>
                                    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24"></div>
                                </div>
                                <div className="text-right">
                                    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16 mb-1"></div>
                                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-12"></div>
                                </div>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className={`bg-card border-0 shadow-md mb-6 rounded-none ${className}`}>
            <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                <div className="flex-1 min-w-0">
                    <CardTitle className="text-base sm:text-lg font-medium text-foreground flex items-center gap-2">
                        Geographic Analytics
                    </CardTitle>
                    <p className="text-xs sm:text-sm text-muted-foreground mt-1">Visitor distribution across locations worldwide</p>
                </div>
                <Tabs value={geoTab} onValueChange={setGeoTab} className="w-full sm:w-auto flex-shrink-0">
                    <TabsList className="grid w-full grid-cols-4 h-9 sm:h-8 gap-1">
                        <TabsTrigger className='text-xs px-1 sm:px-2 md:px-3 truncate' value="map">Map</TabsTrigger>
                        <TabsTrigger className='text-xs px-1 sm:px-2 md:px-3 truncate' value="countries">Countries</TabsTrigger>
                        <TabsTrigger className='text-xs px-1 sm:px-2 md:px-3 truncate' value="cities">Cities</TabsTrigger>
                        <TabsTrigger className='text-xs px-1 sm:px-2 md:px-3 truncate' value="continents">Continents</TabsTrigger>
                    </TabsList>
                </Tabs>
            </CardHeader>
            <CardContent className="p-4 sm:px-6 pt-0">
                <div className="mt-0">
                    {geoTab === 'map' && (
                        <div>
                            {/* Interactive World Map */}
                            <WorldMap
                                data={data?.countries || []}
                                isLoading={isLoading}
                            />
                        </div>
                    )}

                    {geoTab === 'countries' && (
                        <div className="space-y-3">
                            {data?.countries?.slice(0, 15).map((country, index) => {
                                const colors = ['#3B82F6', '#8B5CF6', '#10B981', '#F59E0B', '#EF4444', '#EC4899', '#06B6D4'];
                                const color = colors[index % colors.length];

                                return (
                                    <div key={country.name} className="flex items-center justify-between p-3 border-b hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                        <div className="flex items-center gap-3">
                                            <span className="text-sm font-medium text-gray-500 w-6">#{index + 1}</span>
                                            <div className="relative w-8 h-6 rounded overflow-hidden shadow-sm border border-gray-200 dark:border-gray-600">
                                                <Image
                                                    src={getCountryFlag(country.name)}
                                                    alt={`${country.name} flag`}
                                                    width={32}
                                                    height={24}
                                                    className="object-cover"
                                                    onError={(e) => {
                                                        const target = e.target as HTMLImageElement;
                                                        target.style.display = 'none';
                                                        const fallback = target.parentElement?.querySelector('.flag-fallback') as HTMLElement;
                                                        if (fallback) fallback.style.display = 'flex';
                                                    }}
                                                />
                                                <div className="flag-fallback hidden absolute inset-0 bg-gray-200 dark:bg-gray-700 rounded text-xs font-bold text-gray-600 dark:text-gray-300 items-center justify-center">
                                                    {country.name.substring(0, 2).toUpperCase()}
                                                </div>
                                            </div>
                                            <div>
                                                <p className="font-medium text-sm">{country.name}</p>
                                                <p className="text-xs text-muted-foreground">{country.percentage.toFixed(1)}% of total visitors</p>
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-3">
                                            <div className="text-right">
                                                <p className="text-sm font-semibold">{formatNumber(country.count)}</p>
                                                <p className="text-xs text-muted-foreground">visitors</p>
                                            </div>
                                            <div className="w-16 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all duration-300"
                                                    style={{
                                                        width: `${Math.min(country.percentage, 100)}%`,
                                                        backgroundColor: color
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    {geoTab === 'cities' && (
                        <div className="space-y-3">
                            {data?.cities?.slice(0, 20).map((city, index) => {
                                const colors = ['#EF4444', '#F59E0B', '#10B981', '#3B82F6', '#8B5CF6', '#EC4899', '#06B6D4'];
                                const color = colors[index % colors.length];

                                return (
                                    <div key={city.name} className="flex items-center justify-between p-3 border-b hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                        <div className="flex items-center gap-3">
                                            <span className="text-sm font-medium text-gray-500 w-6">#{index + 1}</span>
                                            <div className="p-2 rounded-lg" style={{ backgroundColor: `${color}20` }}>
                                                <MapPin className="h-4 w-4" style={{ color }} />
                                            </div>
                                            <div>
                                                <p className="font-medium text-sm">{city.name}</p>
                                                <p className="text-xs text-muted-foreground">{city.percentage.toFixed(1)}% of total visitors</p>
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-3">
                                            <div className="text-right">
                                                <p className="text-sm font-semibold">{formatNumber(city.count)}</p>
                                                <p className="text-xs text-muted-foreground">visitors</p>
                                            </div>
                                            <div className="w-16 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all duration-300"
                                                    style={{
                                                        width: `${Math.min(city.percentage, 100)}%`,
                                                        backgroundColor: color
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    {geoTab === 'continents' && (
                        <div className="space-y-3">
                            {data?.continents?.map((continent, index) => {
                                const colors = ['#8B5CF6', '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#EC4899', '#06B6D4'];
                                const color = colors[index % colors.length];

                                return (
                                    <div key={continent.name} className="flex items-center justify-between p-4 bg-white dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700 hover:shadow-sm transition-shadow">
                                        <div className="flex items-center gap-3">
                                            <span className="text-sm font-medium text-gray-500 w-6">#{index + 1}</span>
                                            <div className="p-2 rounded-lg" style={{ backgroundColor: `${color}20` }}>
                                                <div className="text-lg">{getContinentEmoji(continent.name)}</div>
                                            </div>
                                            <div>
                                                <p className="font-medium text-sm">{continent.name}</p>
                                                <p className="text-xs text-muted-foreground">{continent.percentage.toFixed(1)}% of total visitors</p>
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-3">
                                            <div className="text-right">
                                                <p className="text-sm font-semibold">{formatNumber(continent.count)}</p>
                                                <p className="text-xs text-muted-foreground">visitors</p>
                                            </div>
                                            <div className="w-16 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all duration-300"
                                                    style={{
                                                        width: `${Math.min(continent.percentage, 100)}%`,
                                                        backgroundColor: color
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    {/* Empty State */}
                    {!data?.countries?.length && !isLoading && (
                        <div className="text-center py-12 text-muted-foreground">
                            <Globe className="h-16 w-16 mx-auto mb-4 text-gray-300 dark:text-gray-600" />
                            <div className="text-lg font-medium mb-2">No geographic data available</div>
                            <div className="text-sm">Geographic analytics will appear here once visitors start coming to your site</div>
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}