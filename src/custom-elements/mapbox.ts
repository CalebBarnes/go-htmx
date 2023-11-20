import "mapbox-gl/dist/mapbox-gl.css";
import mapboxgl from "mapbox-gl";

type MapboxProps = {
  id: string;
  accessToken: string;
  lat: number;
  lng: number;
  zoom: number;
  controls: boolean;
  mapStyle: string;
  center: [number, number];
  spin: boolean;
};

export function MapboxComponent(
  {
    id = "mapbox-container",
    mapStyle = "mapbox://styles/mapbox/streets-v12",
    center = [0, 0],
    zoom = 5,
    controls = false,
    accessToken,
    lat,
    lng,
    spin = true,
  }: MapboxProps,
  render: (str: string) => void,
) {
  render(
    /*html*/ `<div id="${id}" class="h-full w-full animate-fade-in"></div>`,
  );

  mapboxgl.accessToken = accessToken;
  const map = new mapboxgl.Map({
    container: id, // container ID
    style: mapStyle, // style URL
    center,
    zoom, // starting zoom
    projection: {
      name: "globe",
    },
  });

  // Add markers to the map.
  for (const marker of geojson.features) {
    // Create a DOM element for each marker.
    const el = document.createElement("div");
    const width = marker.properties.iconSize[0];
    const height = marker.properties.iconSize[1];
    el.className = "marker";
    el.style.backgroundImage = `url(https://placekitten.com/g/${width}/${height}/)`;
    el.style.width = `${width}px`;
    el.style.height = `${height}px`;
    el.style.backgroundSize = "100%";

    el.addEventListener("click", () => {
      window.alert(marker.properties.message);
    });

    // Add markers to the map.
    new mapboxgl.Marker(el).setLngLat(marker.geometry.coordinates).addTo(map);
  }

  map.on("style.load", () => {
    map.setFog({}); // Set the default atmosphere style
  });

  // Above zoom level 5, do not rotate.
  const maxSpinZoom = 5;
  // Rotate at intermediate speeds between zoom levels 3 and 5.
  const slowSpinZoom = 3;
  const secondsPerRevolution = 300;

  let userInteracting = false;
  let spinEnabled = spin;

  function spinGlobe() {
    const zoom = map.getZoom();
    if (spinEnabled && !userInteracting && zoom < maxSpinZoom) {
      let distancePerSecond = 360 / secondsPerRevolution;
      if (zoom > slowSpinZoom) {
        // Slow spinning at higher zooms
        const zoomDif = (maxSpinZoom - zoom) / (maxSpinZoom - slowSpinZoom);
        distancePerSecond *= zoomDif;
      }
      const center = map.getCenter();
      center.lng -= distancePerSecond;
      // Smoothly animate the map over one second.
      // When this animation is complete, it calls a 'moveend' event.
      map.easeTo({ center, duration: 1000, easing: (n) => n });
    }
  }

  map.on("load", () => {
    // remove label layers
    const layers = map.getStyle().layers;
    for (const layer of layers) {
      if (layer.type === "symbol" && layer?.layout?.["text-field"]) {
        // remove text labels
        map.removeLayer(layer.id);
      }
    }
  });

  // Pause spinning on interaction
  map.on("mousedown", (e) => {
    userInteracting = true;
  });

  // Restart spinning the globe when interaction is complete
  map.on("mouseup", () => {
    userInteracting = false;
    spinGlobe();
  });

  // These events account for cases where the mouse has moved
  // off the map, so 'mouseup' will not be fired.
  map.on("dragend", () => {
    userInteracting = false;
    spinGlobe();
  });
  map.on("pitchend", () => {
    userInteracting = false;
    spinGlobe();
  });
  map.on("rotateend", () => {
    userInteracting = false;
    spinGlobe();
  });

  // When animation is complete, start spinning if there is no ongoing interaction
  map.on("moveend", () => {
    spinGlobe();
  });

  // add controls
  if (controls) {
    map.addControl(new mapboxgl.NavigationControl());
  }

  // add one marker
  if (lat && lng) {
    new mapboxgl.Marker()
      .setLngLat([lng, lat])
      .setPopup(
        new mapboxgl.Popup().setHTML(/*html*/ `
        <div class="text-black">
          <span>Hey! 👋 What's up?</span>
        </div>`),
      )
      .addTo(map);
  }

  spinGlobe();

  return;
}

const geojson = {
  type: "FeatureCollection",
  features: [
    {
      type: "Feature",
      properties: {
        message: "Foo",
        iconSize: [60, 60],
      },
      geometry: {
        type: "Point",
        coordinates: [-66.324462, -16.024695],
      },
    },
    {
      type: "Feature",
      properties: {
        message: "Bar",
        iconSize: [50, 50],
      },
      geometry: {
        type: "Point",
        coordinates: [-61.21582, -15.971891],
      },
    },
    {
      type: "Feature",
      properties: {
        message: "Baz",
        iconSize: [40, 40],
      },
      geometry: {
        type: "Point",
        coordinates: [-63.292236, -18.281518],
      },
    },
  ],
};
