'use client';

import { ComposableMap, Geographies, Geography, Marker } from 'react-simple-maps';

interface CountryData {
    name: string;
    count: number;
    percentage: number;
}

interface WorldMapProps {
    data?: CountryData[];
    isLoading?: boolean;
    className?: string;
}

// Country coordinates for markers
const countryCoordinates: Record<string, [number, number]> = {
    'United States': [-95.7129, 37.0902],
    'United Kingdom': [-3.4360, 55.3781],
    'Germany': [10.4515, 51.1657],
    'Canada': [-106.3468, 56.1304],
    'France': [2.2137, 46.2276],
    'Australia': [133.7751, -25.2744],
    'Japan': [138.2529, 36.2048],
    'Netherlands': [5.2913, 52.1326],
    'India': [78.9629, 20.5937],
    'Brazil': [-51.9253, -14.2350],
    'Spain': [-3.7492, 40.4637],
    'Italy': [12.5674, 41.8719]
};

// Country ISO codes for geography matching
const countryIsoCodes: Record<string, string> = {
    'United States': 'US',
    'United Kingdom': 'GB',
    'Germany': 'DE',
    'Canada': 'CA',
    'France': 'FR',
    'Australia': 'AU',
    'Japan': 'JP',
    'Netherlands': 'NL',
    'India': 'IN',
    'Brazil': 'BR',
    'Spain': 'ES',
    'Italy': 'IT'
};




export default function WorldMap({ data = [], isLoading = false, className = '' }: WorldMapProps) {
    if (isLoading) {
        return (
            <div className={`h-[32rem] bg-gradient-to-br from-blue-50 to-green-50 dark:from-blue-950/20 dark:to-green-950/20 rounded-lg flex items-center justify-center ${className}`}>
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-sm text-gray-500">Loading geographic data...</p>
                </div>
            </div>
        );
    }

    // If no data, show empty state
    if (!data || data.length === 0) {
        return (
            <div className={`h-[32rem] bg-gradient-to-br from-blue-50 to-green-50 dark:from-blue-950/20 dark:to-green-950/20 rounded-lg flex items-center justify-center ${className}`}>
                <div className="text-center">
                    <div className="text-6xl mb-4">🗺️</div>
                    <h3 className="text-lg font-semibold text-gray-600 dark:text-gray-300 mb-2">No Geographic Data</h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Map will appear when visitor data is available</p>
                </div>
            </div>
        );
    }

    // Get color based on visitor count
    const getCountryColor = (countryName: string): string => {
        const country = data.find(d => d.name === countryName);
        if (!country) return '#E5E7EB'; // Gray for no data

        const maxCount = Math.max(...data.map(d => d.count));
        const intensity = country.count / maxCount;

        if (intensity > 0.8) return '#DC2626'; // Red
        if (intensity > 0.6) return '#EA580C'; // Orange
        if (intensity > 0.4) return '#D97706'; // Amber
        if (intensity > 0.2) return '#059669'; // Emerald
        return '#3B82F6'; // Blue
    };

    // Get marker size based on visitor count
    const getMarkerSize = (count: number): number => {
        const maxCount = Math.max(...data.map(d => d.count));
        const intensity = count / maxCount;
        return Math.max(4, Math.min(20, intensity * 20));
    };

    return (
        <div className={`h-[32rem] rounded-lg overflow-hidden relative ${className}`}>
            {/* Header */}
            {/* <div className="absolute top-0 left-0 right-0 z-10 p-4 bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm border-b border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-gray-800 dark:text-gray-200">World Visitor Map</h3>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                        {data.length} countries
                    </span>
                </div>
            </div> */}

            {/* Map */}
            <div className="h-full pt-4">
                <ComposableMap
                    projection="geoMercator"
                    projectionConfig={{
                        scale: 120,
                        center: [0, 20]
                    }}
                    width={800}
                    height={400}
                    style={{ width: '100%', height: '100%' }}
                >
                    <Geographies geography="https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json">
                        {({ geographies }) =>
                            geographies.map((geo) => {
                                const countryName = Object.keys(countryIsoCodes).find(
                                    name => countryIsoCodes[name] === geo.properties.ISO_A2
                                );

                                return (
                                    <Geography
                                        key={geo.rsmKey}
                                        geography={geo}
                                        fill={getCountryColor(countryName || '')}
                                        stroke="#FFFFFF"
                                        strokeWidth={0.5}
                                        style={{
                                            default: {
                                                outline: 'none',
                                            },
                                            hover: {
                                                fill: '#F59E0B',
                                                outline: 'none',
                                                cursor: 'pointer'
                                            },
                                            pressed: {
                                                outline: 'none',
                                            },
                                        }}
                                    />
                                );
                            })
                        }
                    </Geographies>

                    {/* Markers for countries with data */}
                    {data
                        .filter(country => countryCoordinates[country.name])
                        .map((country) => (
                            <Marker
                                key={country.name}
                                coordinates={countryCoordinates[country.name]}
                            >
                                <circle
                                    r={getMarkerSize(country.count)}
                                    fill={getCountryColor(country.name)}
                                    stroke="#FFFFFF"
                                    strokeWidth={2}
                                    style={{ cursor: 'pointer' }}
                                />
                                <title>
                                    {country.name}: {country.count.toLocaleString()} visitors ({country.percentage.toFixed(1)}%)
                                </title>
                            </Marker>
                        ))}
                </ComposableMap>
            </div>

            {/* Legend */}
            <div className="absolute bottom-4 left-4 bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-lg p-3 z-10">
                <h4 className="text-sm font-semibold mb-2 text-gray-800 dark:text-gray-200">Visitor Distribution</h4>
                <div className="space-y-1">
                    <div className="flex items-center gap-2 text-xs">
                        <div className="w-3 h-3 rounded-full bg-red-600"></div>
                        <span className="text-gray-600 dark:text-gray-400">High (80%+)</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                        <div className="w-3 h-3 rounded-full bg-orange-600"></div>
                        <span className="text-gray-600 dark:text-gray-400">Medium-High (60-80%)</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                        <div className="w-3 h-3 rounded-full bg-amber-600"></div>
                        <span className="text-gray-600 dark:text-gray-400">Medium (40-60%)</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                        <div className="w-3 h-3 rounded-full bg-emerald-600"></div>
                        <span className="text-gray-600 dark:text-gray-400">Low-Medium (20-40%)</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                        <div className="w-3 h-3 rounded-full bg-blue-600"></div>
                        <span className="text-gray-600 dark:text-gray-400">Low (0-20%)</span>
                    </div>
                </div>
            </div>

            {/* Stats */}
            <div className="absolute bottom-4 right-4 bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-lg p-3 z-10">
                <div className="text-xs space-y-1">
                    <div className="text-gray-600 dark:text-gray-400">Total Visitors:</div>
                    <div className="font-semibold text-gray-800 dark:text-gray-200">
                        {data.reduce((sum, country) => sum + country.count, 0).toLocaleString()}
                    </div>
                </div>
            </div>
        </div>
    );
}
