import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";

export class MapBoxElement extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.innerHTML = /*html*/ `
    <div id="map-container" class="min-h-[500px]"></div>`;

    const accessToken = this.getAttribute("accessToken") as string;
    mapboxgl.accessToken = accessToken;

    const map = new mapboxgl.Map({
      container: "map-container", // container ID
      // Choose from Mapbox's core styles, or make your own style with Mapbox Studio
      style: "mapbox://styles/mapbox/streets-v12", // style URL
      center: [-74.5, 40], // starting position [lng, lat]
      zoom: 9, // starting zoom
    });

    // add controls
    map.addControl(new mapboxgl.NavigationControl());

    const markerLocation = {
      lng: -74.5,
      lat: 40,
    };

    // add markers
    new mapboxgl.Marker()
      .setLngLat([markerLocation.lng, markerLocation.lat])
      .setPopup(
        new mapboxgl.Popup().setHTML(/*html*/ `
          <div class="text-black">
            <span>Tooltip: lat:${markerLocation.lat} lng:${markerLocation.lng}</span>
          </div>`),
      )
      .addTo(map);
  }
}
