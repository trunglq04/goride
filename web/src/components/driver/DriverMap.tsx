'use client';

import { useDriverStreamConnection } from '@/hooks/useDriverStreamConnection';
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet';
import L from 'leaflet';
import { MapClickHandler } from '@/components/map/MapClickHandler';
import { useMemo, useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { CarPackageSlug, Coordinate } from '@/types';
import { DriverTripOverview } from './DriverTripOverview';
import * as Geohash from 'ngeohash';
import { RoutingControl } from '@/components/map/RoutingControl';
import { DriverCard } from './DriverCard';
import { TripEvents } from '@/types/contracts';
import { uuid } from '@/utils/uuid';
import { VITE_CARTO_API_KEY } from '@/config/constants';
import { useAuth } from '@/contexts/AuthContext';
import { AuthRequiredModal } from '@/components/auth/AuthRequiredModal';
import {
  pickupMarkerIcon,
  destinationMarkerIcon,
  driverMarkerIcon,
} from '@/components/map/MapMarkers';
import { Lock, AlertCircle } from 'lucide-react';

const START_LOCATION: Coordinate = {
  latitude: 37.7749,
  longitude: -122.4194,
};

export const DriverMap = ({ packageSlug }: { packageSlug: CarPackageSlug }) => {
  const router = useRouter();
  const { user, accessToken, isAuthenticated } = useAuth();

  const [showAuthModal, setShowAuthModal] = useState(!isAuthenticated);
  const mapRef = useRef<L.Map>(null);
  const guestUserID = useMemo(() => uuid(), []);
  const userID = user?.id || guestUserID;

  const [riderLocation, setRiderLocation] =
    useState<Coordinate>(START_LOCATION);

  const driverGeohash = useMemo(
    () => Geohash.encode(riderLocation?.latitude, riderLocation?.longitude, 7),
    [riderLocation?.latitude, riderLocation?.longitude]
  );

  const {
    error,
    driver,
    tripStatus,
    requestedTrip,
    sendMessage,
    setTripStatus,
    resetTripStatus,
  } = useDriverStreamConnection({
    location: riderLocation,
    geohash: driverGeohash,
    userID,
    packageSlug,
  });

  const handleMapClick = (e: L.LeafletMouseEvent) => {
    if (!isAuthenticated || !accessToken) {
      setShowAuthModal(true);
      return;
    }

    const newLocation = {
      latitude: e.latlng.lat,
      longitude: e.latlng.lng,
    };
    setRiderLocation(newLocation);

    const newGeohash = Geohash.encode(
      newLocation.latitude,
      newLocation.longitude,
      7
    );

    sendMessage({
      type: TripEvents.DriverLocation,
      data: {
        driverID: userID,
        location: newLocation,
        geohash: newGeohash,
      },
    });
  };

  const handleAcceptTrip = () => {
    if (!isAuthenticated || !accessToken) {
      setShowAuthModal(true);
      return;
    }

    if (!requestedTrip || !requestedTrip.id || !driver) {
      alert('No trip ID found or driver is not set');
      return;
    }

    sendMessage({
      type: TripEvents.DriverTripAccept,
      data: {
        tripID: requestedTrip.id,
        riderID: requestedTrip.userID,
        driver: driver,
      },
    });

    setTripStatus(TripEvents.DriverTripAccept);
  };

  const handleDeclineTrip = () => {
    if (!requestedTrip || !requestedTrip.id || !driver) {
      alert('No trip ID found or driver is not set');
      return;
    }

    sendMessage({
      type: TripEvents.DriverTripDecline,
      data: {
        tripID: requestedTrip.id,
        riderID: requestedTrip.userID,
        driver: driver,
      },
    });

    setTripStatus(TripEvents.DriverTripDecline);
    resetTripStatus();
  };

  const parsedRoute = useMemo(
    () =>
      requestedTrip?.route?.geometry[0]?.coordinates.map(
        (coord) => [coord?.longitude, coord?.latitude] as [number, number]
      ),
    [requestedTrip]
  );

  const destination = useMemo(
    () =>
      requestedTrip?.route?.geometry[0]?.coordinates[
        requestedTrip?.route?.geometry[0]?.coordinates?.length - 1
      ],
    [requestedTrip]
  );

  const startLocation = useMemo(
    () => requestedTrip?.route?.geometry[0]?.coordinates[0],
    [requestedTrip]
  );

  if (error) {
    return (
      <div className="p-8 text-center text-red-600 bg-red-50 rounded-2xl m-4">
        Driver stream connection error: {error}
      </div>
    );
  }

  return (
    <div className="relative flex flex-col md:flex-row h-[calc(100vh-64px)] overflow-hidden">
      {/* Driver Auth Required Banner if not signed in */}
      {!isAuthenticated && (
        <div className="absolute top-3 left-4 right-4 md:right-[420px] max-w-md mx-auto z-30 pointer-events-auto">
          <div className="p-2.5 bg-amber-50/95 backdrop-blur-md border border-amber-200/80 rounded-2xl shadow-sm flex items-center justify-between gap-3 text-xs text-amber-950">
            <div className="flex items-center gap-2 truncate">
              <span className="w-2 h-2 rounded-full bg-amber-600 animate-pulse shrink-0" />
              <span className="font-semibold truncate">
                Driver Sign In Required for Telemetry
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

      {/* Main Map */}
      <div className="flex-1 relative h-full">
        <MapContainer
          center={[riderLocation.latitude, riderLocation.longitude]}
          zoom={13}
          style={{ height: '100%', width: '100%' }}
          ref={mapRef}
          zoomControl={false}
        >
          <TileLayer
            url={`https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png?key=${VITE_CARTO_API_KEY}`}
            attribution="&copy; OpenStreetMap contributors &copy; CARTO"
          />

          <Marker
            key={userID}
            position={[riderLocation.latitude, riderLocation.longitude]}
            icon={driverMarkerIcon}
          >
            <Popup>
              <div className="text-xs p-1">
                <p className="font-bold text-[#18181b]">Driver Location</p>
                <p className="text-zinc-500 font-mono text-[10px]">
                  Geohash: {driverGeohash}
                </p>
              </div>
            </Popup>
          </Marker>

          {startLocation && (
            <Marker
              position={[startLocation.longitude, startLocation.latitude]}
              icon={pickupMarkerIcon}
            >
              <Popup>
                <div className="text-xs font-semibold p-1">Rider Pickup</div>
              </Popup>
            </Marker>
          )}

          {destination && (
            <Marker
              position={[destination.longitude, destination.latitude]}
              icon={destinationMarkerIcon}
            >
              <Popup>
                <div className="text-xs font-semibold p-1">Destination</div>
              </Popup>
            </Marker>
          )}

          {parsedRoute && <RoutingControl route={parsedRoute} />}
          <MapClickHandler onClick={handleMapClick} />
        </MapContainer>
      </div>

      {/* Driver Sidebar Console (Tactile Card Column) */}
      <div className="flex flex-col md:w-[380px] bg-white border-t md:border-t-0 md:border-l border-black/8 z-10 shadow-sm">
        <div className="p-4 border-b border-black/5">
          <DriverCard driver={driver} packageSlug={packageSlug} />
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          <DriverTripOverview
            trip={requestedTrip}
            status={tripStatus}
            onAcceptTrip={handleAcceptTrip}
            onDeclineTrip={handleDeclineTrip}
            onDismiss={resetTripStatus}
          />
        </div>
      </div>

      {/* Auth Required Modal Dialog */}
      <AuthRequiredModal
        isOpen={showAuthModal}
        onClose={() => setShowAuthModal(false)}
        role="driver"
        errorMessage="You must sign in to an authorized Driver Partner account to broadcast live vehicle GPS coordinates and receive dispatches."
      />
    </div>
  );
};
