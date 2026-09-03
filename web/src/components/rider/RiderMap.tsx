'use client';

import { useRiderStreamConnection } from '@/hooks/useRiderStreamConnection';
import {
  MapContainer,
  Marker,
  Popup,
  Rectangle,
  TileLayer,
} from 'react-leaflet';
import L from 'leaflet';
import { getGeohashBounds } from '@/utils/geohash';
import { uuid } from '@/utils/uuid';
import { useMemo, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { MapClickHandler } from '@/components/map/MapClickHandler';
import {
  RouteFare,
  RequestRideProps,
  TripPreview,
  HTTPTripStartResponse,
  Coordinate,
} from '@/types';
import { RoutingControl } from '@/components/map/RoutingControl';
import { VITE_CARTO_API_KEY } from '@/config/constants';
import { RiderTripOverview } from './RiderTripOverview';
import { TripEvents } from '@/types/contracts';
import {
  pickupMarkerIcon,
  destinationMarkerIcon,
  driverMarkerIcon,
} from '@/components/map/MapMarkers';
import { MapPin, Lock } from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { AuthRequiredModal } from '@/components/auth/AuthRequiredModal';
import { tripService } from '@/services/tripService';
import { ApiError } from '@/services/apiClient';

interface RiderMapProps {
  onRouteSelected?: (distance: number) => void;
}

export default function RiderMap({ onRouteSelected }: RiderMapProps) {
  const router = useRouter();
  const { user, accessToken, isAuthenticated } = useAuth();

  const [trip, setTrip] = useState<TripPreview | null>(null);
  const [destination, setDestination] = useState<[number, number] | null>(null);
  const [destinationName, setDestinationName] = useState<string>('');
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [authErrorMessage, setAuthErrorMessage] = useState<string | null>(null);

  const mapRef = useRef<L.Map>(null);
  const guestUserID = useMemo(() => uuid(), []);
  const activeUserID = user?.id || guestUserID;
  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const location = {
    latitude: 37.7749,
    longitude: -122.4194,
  };

  const {
    drivers,
    error,
    tripStatus,
    assignedDriver,
    paymentSession,
    resetTripStatus,
    setTripStatus,
  } = useRiderStreamConnection(location, activeUserID);

  const triggerPreview = async (
    destLat: number,
    destLng: number,
    name?: string
  ) => {
    setDestination([destLat, destLng]);
    if (name) setDestinationName(name);

    if (!isAuthenticated || !accessToken) {
      setAuthErrorMessage(
        'You must sign in to calculate route fares and book rides.'
      );
      setShowAuthModal(true);
      return;
    }

    try {
      const data = await tripService.previewTrip(
        {
          userID: activeUserID,
          pickup: {
            latitude: location.latitude,
            longitude: location.longitude,
          },
          destination: {
            latitude: destLat,
            longitude: destLng,
          },
        },
        accessToken
      );

      const parsedRoute = data.route.geometry[0].coordinates.map(
        (coord: Coordinate) => [coord.longitude, coord.latitude] as [number, number]
      );

      setTrip({
        tripID: '',
        route: parsedRoute,
        rideFares: data.rideFares,
        distance: data.route.distance,
        duration: data.route.duration,
      });

      onRouteSelected?.(data.route.distance);
    } catch (err: any) {
      console.error('Failed to preview ride', err);
      if (err instanceof ApiError && (err.status === 401 || err.message.includes('Authorization'))) {
        setAuthErrorMessage('Please sign in to your GoRide account.');
        setShowAuthModal(true);
      }
    }
  };

  const handleMapClick = async (e: L.LeafletMouseEvent) => {
    if (trip?.tripID) {
      return;
    }

    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current);
    }

    debounceTimeoutRef.current = setTimeout(async () => {
      await triggerPreview(
        e.latlng.lat,
        e.latlng.lng,
        `Pin (${e.latlng.lat.toFixed(3)}, ${e.latlng.lng.toFixed(3)})`
      );
    }, 400);
  };

  const handleStartTrip = async (fare: RouteFare) => {
    if (!isAuthenticated || !accessToken) {
      setAuthErrorMessage('Please sign in to confirm and book this ride.');
      setShowAuthModal(true);
      return;
    }

    if (!fare.id) {
      alert('No Fare ID in the payload');
      return;
    }

    try {
      const data = await tripService.startTrip(
        {
          rideFareID: fare.id,
          userID: activeUserID,
        },
        accessToken
      );

      if (trip) {
        setTrip(
          (prev) =>
            ({
              ...prev,
              tripID: data.tripID,
            }) as TripPreview
        );
        setTripStatus(TripEvents.Created);
      }

      return data;
    } catch (err: any) {
      console.error('Start trip failed:', err);
      if (err instanceof ApiError && (err.status === 401 || err.message.includes('Authorization'))) {
        setAuthErrorMessage('Please sign in to confirm your ride.');
        setShowAuthModal(true);
        return;
      }
      alert('Failed to start trip. Please try again.');
    }
  };

  const handleCancelTrip = async () => {
    if (trip && trip.tripID) {
      try {
        await tripService.cancelTrip(
          {
            tripID: trip.tripID,
            userID: activeUserID,
            reason: 'User requested cancellation',
          },
          accessToken
        );
      } catch (err) {
        console.error('Failed to cancel trip on backend', err);
      }
    }
    setTrip(null);
    setDestination(null);
    setDestinationName('');
    resetTripStatus();
  };

  if (error) {
    return (
      <div className="p-8 text-center text-red-600 bg-red-50 rounded-2xl m-4">
        Connection error: {error}
      </div>
    );
  }

  return (
    <div className="relative w-full h-[calc(100vh-64px)] overflow-hidden bg-[#eef0f3]">
      {/* 0. Guest Mode Sign-In Required Header Banner */}
      {!isAuthenticated && (
        <div className="absolute top-2 left-4 right-4 max-w-md mx-auto z-30 pointer-events-auto">
          <div className="p-2.5 bg-amber-50/95 backdrop-blur-md border border-amber-200/80 rounded-2xl shadow-sm flex items-center justify-between gap-3 text-xs text-amber-950">
            <div className="flex items-center gap-2 truncate">
              <span className="w-2 h-2 rounded-full bg-amber-600 animate-pulse shrink-0" />
              <span className="font-semibold truncate">
                Sign In Required to Request Rides
              </span>
            </div>
            <div className="flex items-center gap-1.5 shrink-0">
              <button
                type="button"
                onClick={() => router.push('/login')}
                className="px-2.5 py-1 rounded-lg bg-[#18181b] text-white font-semibold text-[11px] tactile-press cursor-pointer"
              >
                Sign In
              </button>
              <button
                type="button"
                onClick={() => router.push('/register')}
                className="px-2.5 py-1 rounded-lg bg-white border border-black/10 text-zinc-800 font-semibold text-[11px] tactile-press cursor-pointer"
              >
                Register
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 1. Full-Screen Ambient Map Canvas */}
      <div className="absolute inset-0 w-full h-full z-0">
        <MapContainer
          center={[location.latitude, location.longitude]}
          zoom={14}
          style={{ height: '100%', width: '100%' }}
          ref={mapRef}
          zoomControl={false}
        >
          <TileLayer
            url={`https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png?key=${VITE_CARTO_API_KEY}`}
            attribution="&copy; OpenStreetMap contributors &copy; CARTO"
          />

          {/* User Origin Pin */}
          <Marker
            position={[location.latitude, location.longitude]}
            icon={pickupMarkerIcon}
          >
            <Popup>
              <div className="text-xs font-semibold p-1">
                Pickup: Current Location
              </div>
            </Popup>
          </Marker>

          {/* Geohash broadcast cells */}
          {drivers?.map((driver) => (
            <Rectangle
              key={`grid-${driver?.geohash}`}
              bounds={
                getGeohashBounds(driver?.geohash) as L.LatLngBoundsExpression
              }
              pathOptions={{
                color: '#18181b',
                weight: 0.75,
                fillOpacity: 0.04,
              }}
            />
          ))}

          {/* Real-time Driver Markers (Top-Down Directional SVGs) */}
          {drivers?.map((driver) => (
            <Marker
              key={driver?.id}
              position={[
                driver?.location?.latitude,
                driver?.location?.longitude,
              ]}
              icon={driverMarkerIcon}
            >
              <Popup>
                <div className="text-xs p-1">
                  <p className="font-bold text-[#18181b]">{driver?.name}</p>
                  <p className="text-zinc-500 font-mono mt-0.5">
                    Plate: {driver?.carPlate?.toUpperCase()}
                  </p>
                </div>
              </Popup>
            </Marker>
          ))}

          {/* Destination Pin */}
          {destination && (
            <Marker position={destination} icon={destinationMarkerIcon}>
              <Popup>
                <div className="text-xs font-semibold p-1">
                  Destination: {destinationName || 'Selected Point'}
                </div>
              </Popup>
            </Marker>
          )}

          {trip && <RoutingControl route={trip.route} />}
          <MapClickHandler onClick={handleMapClick} />
        </MapContainer>
      </div>

      {/* 2. Floating Top Location Island */}
      <div
        className={`absolute ${!isAuthenticated ? 'top-14' : 'top-4'} left-4 right-4 max-w-md mx-auto z-20 pointer-events-auto transition-all`}
      >
        <div className="tactile-surface rounded-2xl p-3.5 flex flex-col gap-2.5">
          <div className="flex items-center justify-between px-1">
            <span className="font-brand text-sm font-bold text-[#18181b]">
              Where would you like to travel?
            </span>
            <span className="text-[11px] font-mono text-zinc-400">
              SF Bay Area
            </span>
          </div>

          <div className="flex flex-col gap-1.5">
            {/* Origin Row */}
            <div className="flex items-center gap-2.5 px-3 py-2 bg-[#fafafb] border border-black/5 rounded-xl text-xs">
              <div className="w-2 h-2 rounded-full bg-[#18181b] shrink-0" />
              <span className="text-zinc-600 font-medium truncate">
                Current Location • Market & 4th St
              </span>
            </div>

            {/* Destination Row */}
            <div className="flex items-center gap-2.5 px-3 py-2 bg-[#ffffff] border border-black/10 rounded-xl text-xs shadow-sm">
              <MapPin className="w-3.5 h-3.5 text-emerald-600 shrink-0" />
              <span className="text-[#18181b] font-semibold truncate">
                {destinationName ||
                  (destination
                    ? 'Destination pinned on map'
                    : 'Tap map or choose preset below')}
              </span>
            </div>
          </div>

          {/* Quick Presets */}
          {!trip && (
            <div className="flex items-center gap-1.5 pt-1 overflow-x-auto">
              <span className="text-[11px] text-zinc-400 font-medium px-1">
                Quick:
              </span>
              <button
                type="button"
                onClick={() =>
                  triggerPreview(37.768, -122.3877, 'Chase Center')
                }
                className="px-2.5 py-1 rounded-lg bg-zinc-100 hover:bg-zinc-200 text-[11px] font-semibold text-zinc-700 transition-colors shrink-0 cursor-pointer"
              >
                Chase Center
              </button>
              <button
                type="button"
                onClick={() => triggerPreview(37.6213, -122.379, 'SFO Airport')}
                className="px-2.5 py-1 rounded-lg bg-zinc-100 hover:bg-zinc-200 text-[11px] font-semibold text-zinc-700 transition-colors shrink-0 cursor-pointer"
              >
                SFO Airport
              </button>
              <button
                type="button"
                onClick={() =>
                  triggerPreview(37.7955, -122.3937, 'Ferry Building')
                }
                className="px-2.5 py-1 rounded-lg bg-zinc-100 hover:bg-zinc-200 text-[11px] font-semibold text-zinc-700 transition-colors shrink-0 cursor-pointer"
              >
                Ferry Building
              </button>
            </div>
          )}
        </div>
      </div>

      {/* 3. Ergonomic Bottom Drawer (Tactile Card Over Map) */}
      <div className="absolute bottom-4 left-4 right-4 max-w-md mx-auto z-20 pointer-events-auto max-h-[65vh] flex flex-col justify-end">
        <RiderTripOverview
          trip={trip}
          assignedDriver={assignedDriver}
          status={tripStatus}
          paymentSession={paymentSession}
          onPackageSelect={handleStartTrip}
          onCancel={handleCancelTrip}
        />
      </div>

      {/* 4. Auth Required Modal */}
      <AuthRequiredModal
        isOpen={showAuthModal}
        onClose={() => setShowAuthModal(false)}
        role="rider"
        errorMessage={authErrorMessage}
      />
    </div>
  );
}
