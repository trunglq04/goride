import { Polyline } from "react-leaflet";

export function RoutingControl({ route }: { route: [number, number][] }) {
  if (!route) return null;

  return (
    <Polyline
      positions={route}
      pathOptions={{
        color: "#18181b",
        weight: 4,
        opacity: 0.85,
        lineCap: "round",
        lineJoin: "round",
      }}
    />
  );
}
