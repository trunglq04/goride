import L from "leaflet";

// Tactile Pickup Origin Marker (Ink circle with central beacon)
const pickupSvg = `
<svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 36 36" fill="none">
  <filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">
    <feDropShadow dx="0" dy="3" stdDeviation="3" flood-color="#000000" flood-opacity="0.25"/>
  </filter>
  <circle cx="18" cy="18" r="14" fill="#FFFFFF" filter="url(#shadow)" stroke="#E4E4E7" stroke-width="1"/>
  <circle cx="18" cy="18" r="10" fill="#18181B"/>
  <circle cx="18" cy="18" r="4" fill="#FFFFFF"/>
</svg>
`;

export const pickupMarkerIcon = new L.Icon({
  iconUrl: `data:image/svg+xml;utf8,${encodeURIComponent(pickupSvg)}`,
  iconSize: [36, 36],
  iconAnchor: [18, 18],
  popupAnchor: [0, -18],
});

// Tactile Destination Arrival Pin (Emerald Beacon)
const destinationSvg = `
<svg xmlns="http://www.w3.org/2000/svg" width="36" height="44" viewBox="0 0 36 44" fill="none">
  <filter id="destShadow" x="-20%" y="-20%" width="140%" height="140%">
    <feDropShadow dx="0" dy="4" stdDeviation="3" flood-color="#000000" flood-opacity="0.3"/>
  </filter>
  <path d="M18 0C8.06 0 0 8.06 0 18C0 30.5 18 44 18 44C18 44 36 30.5 36 18C36 8.06 27.94 0 18 0Z" fill="#059669" filter="url(#destShadow)"/>
  <circle cx="18" cy="18" r="7" fill="#FFFFFF"/>
  <circle cx="18" cy="3" r="3" fill="#059669"/>
</svg>
`;

export const destinationMarkerIcon = new L.Icon({
  iconUrl: `data:image/svg+xml;utf8,${encodeURIComponent(destinationSvg)}`,
  iconSize: [36, 44],
  iconAnchor: [18, 44],
  popupAnchor: [0, -44],
});

// Sleek Top-Down Driver Vehicle Marker (Directional, Local SVG, no external CDN)
const driverCarSvg = `
<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 60 60" fill="none">
  <filter id="carShadow" x="-20%" y="-20%" width="140%" height="140%">
    <feDropShadow dx="0" dy="3" stdDeviation="4" flood-color="#000000" flood-opacity="0.25"/>
  </filter>
  <circle cx="30" cy="30" r="26" fill="#FFFFFF" stroke="#E4E4E7" stroke-width="1.5" filter="url(#carShadow)"/>
  <!-- Top Down Vehicle Body -->
  <g transform="translate(20, 12)">
    <!-- Chassis -->
    <rect x="2" y="2" width="16" height="32" rx="5" fill="#18181B"/>
    <!-- Windshield front -->
    <path d="M4 11 Q10 8 16 11 L15 16 H5 Z" fill="#FFFFFF" opacity="0.3"/>
    <!-- Rear glass -->
    <path d="M5 24 H15 L16 28 Q10 29 4 28 Z" fill="#FFFFFF" opacity="0.3"/>
    <!-- Roof -->
    <rect x="5" y="16" width="10" height="8" rx="1.5" fill="#27272A"/>
    <!-- Headlights -->
    <circle cx="4" cy="3" r="1.5" fill="#FBBF24"/>
    <circle cx="16" cy="3" r="1.5" fill="#FBBF24"/>
  </g>
</svg>
`;

export const driverMarkerIcon = new L.Icon({
  iconUrl: `data:image/svg+xml;utf8,${encodeURIComponent(driverCarSvg)}`,
  iconSize: [40, 40],
  iconAnchor: [20, 20],
  popupAnchor: [0, -20],
});
