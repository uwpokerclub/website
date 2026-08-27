import { Gallery as PhotoGallery, Item } from "react-photoswipe-gallery";
import "photoswipe/dist/photoswipe.css";
import { images } from "./galleryImages";

export function Gallery() {
  return (
    <PhotoGallery>
      {images.map((image) => (
        <Item<HTMLImageElement> key={image.src} original={image.original} width={image.width} height={image.height}>
          {({ ref, open }) => <img ref={ref} onClick={open} src={image.src} />}
        </Item>
      ))}
    </PhotoGallery>
  );
}
