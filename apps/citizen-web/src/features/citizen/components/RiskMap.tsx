import { useEffect, useRef } from "react";
import { Box } from "@mui/material";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

export type RiskMapMarker = {
  lat: number;
  lng: number;
  title: string;
  /** Marker fill colour (from @nadaa/brand severity/hazard roles). */
  color: string;
  /** Small glyph shown inside the marker, e.g. a shelter number. */
  glyph?: string;
  kind?: "risk" | "shelter";
};

type RiskMapProps = {
  lat: number;
  lng: number;
  /** Colour of the primary risk pin (severity foreground). */
  riskColor: string;
  markers?: RiskMapMarker[];
  ariaLabel?: string;
};

function markerIcon(marker: RiskMapMarker) {
  const isRisk = marker.kind !== "shelter";
  const className = isRisk ? "citizen-map-pin citizen-map-pin--risk" : "citizen-map-pin";
  const size = isRisk ? 30 : 26;
  return L.divIcon({
    className: "citizen-map-icon",
    html: `<span class="${className}" style="--marker:${marker.color}">${
      marker.glyph ?? ""
    }</span>`,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  });
}

/**
 * Brand-themed Leaflet mini-map. Tiles are tinted navy via CSS, the
 * attribution and popups are re-skinned to tokens, and markers are
 * severity-coloured divIcons (no default marker assets — so it renders
 * cleanly offline, with the navy frame standing in for missing tiles).
 */
export function RiskMap({
  lat,
  lng,
  riskColor,
  markers = [],
  ariaLabel = "Map of the selected area with risk and shelter markers",
}: RiskMapProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [lat, lng],
      zoom: 13,
      zoomControl: false,
      scrollWheelZoom: false,
      dragging: true,
      doubleClickZoom: false,
      boxZoom: false,
      keyboard: false,
      attributionControl: true,
    });

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
      maxZoom: 19,
    }).addTo(map);

    L.control.zoom({ position: "bottomright" }).addTo(map);

    mapRef.current = map;
    layerRef.current = L.layerGroup().addTo(map);

    return () => {
      map.remove();
      mapRef.current = null;
      layerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const layer = layerRef.current;
    if (!map || !layer) {
      return;
    }

    map.setView([lat, lng], map.getZoom(), { animate: false });
    layer.clearLayers();

    const riskMarker: RiskMapMarker = {
      lat,
      lng,
      title: "Selected area",
      color: riskColor,
      kind: "risk",
    };

    [riskMarker, ...markers].forEach((marker) => {
      L.marker([marker.lat, marker.lng], {
        icon: markerIcon(marker),
        title: marker.title,
        keyboard: false,
      })
        .bindPopup(marker.title)
        .addTo(layer);
    });
  }, [lat, lng, riskColor, markers]);

  return (
    <Box
      ref={containerRef}
      className="citizen-map"
      role="img"
      aria-label={ariaLabel}
    />
  );
}

export default RiskMap;
