import { reactive } from 'vue';

export interface DraftProjectState {
  meta: {
    titleRu: string;
    titleEn: string;
    seoRu: string;
    seoEn: string;
    aboutRu: string;
    aboutEn: string;
  };
  media: {
    icon: { file?: File; preview?: string } | null;
    coverMain: { file?: File; preview?: string } | null;
    video: { file?: File; preview?: string } | null;
  };
  builds: Array<{ version: string; date: string }>;
  activeBuildVersion: string | null;
}

export const draftProject = reactive<DraftProjectState>({
  meta: {
    titleRu: '',
    titleEn: '',
    seoRu: '',
    seoEn: '',
    aboutRu: '',
    aboutEn: '',
  },
  media: {
    icon: null,
    coverMain: null,
    video: null,
  },
  builds: [],
  activeBuildVersion: null,
});

export function resetDraftState(): void {
  draftProject.meta.titleRu = '';
  draftProject.meta.titleEn = '';
  draftProject.meta.seoRu = '';
  draftProject.meta.seoEn = '';
  draftProject.meta.aboutRu = '';
  draftProject.meta.aboutEn = '';
  draftProject.media.icon = null;
  draftProject.media.coverMain = null;
  draftProject.media.video = null;
  draftProject.builds = [];
  draftProject.activeBuildVersion = null;
}
