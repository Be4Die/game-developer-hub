/**
 * mediaCompressor.js — валидация медиа-файлов и легковесная клиентская компрессия скриншотов.
 */

const MAX_IMAGE_SIZE = 15 * 1024 * 1024; // 15 МБ
const MAX_VIDEO_SIZE = 50 * 1024 * 1024; // 50 МБ

const ALLOWED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp'];
const ALLOWED_VIDEO_TYPES = ['video/mp4', 'video/webm'];

/**
 * Проверяет файл на соответствие форматам и лимитам размера.
 * @param {File} file
 * @returns {{ valid: boolean, error?: string, isImage: boolean, isVideo: boolean }}
 */
export function validateChatFile(file) {
  if (!file) {
    return { valid: false, error: 'Файл не выбран', isImage: false, isVideo: false };
  }

  const type = (file.type || '').toLowerCase();
  const name = (file.name || '').toLowerCase();

  const isImage =
    ALLOWED_IMAGE_TYPES.includes(type) ||
    name.endsWith('.png') ||
    name.endsWith('.jpg') ||
    name.endsWith('.jpeg') ||
    name.endsWith('.webp');

  const isVideo =
    ALLOWED_VIDEO_TYPES.includes(type) ||
    name.endsWith('.mp4') ||
    name.endsWith('.webm');

  if (!isImage && !isVideo) {
    return {
      valid: false,
      error: 'Поддерживаются только изображения (PNG, JPEG, WebP) и видео (MP4, WebM)',
      isImage: false,
      isVideo: false,
    };
  }

  if (isImage && file.size > MAX_IMAGE_SIZE) {
    return {
      valid: false,
      error: `Размер изображения превышает 15 МБ (${formatBytes(file.size)})`,
      isImage: true,
      isVideo: false,
    };
  }

  if (isVideo && file.size > MAX_VIDEO_SIZE) {
    return {
      valid: false,
      error: `Размер видео превышает 50 МБ (${formatBytes(file.size)}). Пожалуйста, сожмите видео или обрежьте нужный фрагмент`,
      isImage: false,
      isVideo: true,
    };
  }

  return { valid: true, isImage, isVideo };
}

/**
 * Сжимает тяжелые изображения на клиенте через HTML5 Canvas перед отправкой на сервер.
 * Если файл меньше 600 КБ или сжатие не дало выигрыша — возвращает исходный файл.
 * @param {File} file
 * @param {number} maxDimension макс. ширина или высота в пикселях (по умолчанию 1920)
 * @param {number} quality качество WebP/JPEG (0.85)
 * @returns {Promise<File>}
 */
export async function compressImageIfNeeded(file, maxDimension = 1920, quality = 0.85) {
  if (!file || !file.type.startsWith('image/')) {
    return file;
  }

  // Не сжимаем легкие файлы
  if (file.size < 600 * 1024) {
    return file;
  }

  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onerror = () => resolve(file);
    reader.onload = (e) => {
      const img = new Image();
      img.onerror = () => resolve(file);
      img.onload = () => {
        let { width, height } = img;

        // Масштабирование с сохранением пропорций
        if (width > maxDimension || height > maxDimension) {
          if (width > height) {
            height = Math.round((height * maxDimension) / width);
            width = maxDimension;
          } else {
            width = Math.round((width * maxDimension) / height);
            height = maxDimension;
          }
        } else if (file.size < 1.5 * 1024 * 1024) {
          // Разрешение небольшое и размер умеренный
          resolve(file);
          return;
        }

        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          resolve(file);
          return;
        }

        ctx.drawImage(img, 0, 0, width, height);

        // Предпочитаем формат WebP, фоллбэк на JPEG
        const outputMime = 'image/webp';
        canvas.toBlob(
          (blob) => {
            if (!blob || blob.size >= file.size) {
              resolve(file);
              return;
            }

            const baseName = file.name.replace(/\.[^/.]+$/, '');
            const compressedFile = new File([blob], `${baseName}.webp`, {
              type: outputMime,
              lastModified: Date.now(),
            });
            resolve(compressedFile);
          },
          outputMime,
          quality
        );
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  });
}

/**
 * Форматирует размер файла в удобочитаемый вид
 * @param {number} bytes
 * @returns {string}
 */
export function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0\u00A0Б';
  const k = 1024;
  const sizes = ['Б', 'КБ', 'МБ', 'ГБ'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))}\u00A0${sizes[i]}`;
}
