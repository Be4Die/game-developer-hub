import { reactive } from 'vue';

export const draftProject = reactive({
  meta: {
    titleRu: '',
    titleEn: '',
    seoRu: '',
    seoEn: '',
    about: '',
  },
  media: {
    icon: null, // { file, preview }
    coverMain: null, // { file, preview }
    video: null, // { file, preview }
  },
  builds: [], // { version: string, date: string }
  activeBuildVersion: null,
});

export function resetDraftState() {
  draftProject.meta.titleRu = '';
  draftProject.meta.titleEn = '';
  draftProject.meta.seoRu = '';
  draftProject.meta.seoEn = '';
  draftProject.meta.about = '';
  draftProject.media.icon = null;
  draftProject.media.coverMain = null;
  draftProject.media.video = null;
  draftProject.builds = [];
  draftProject.activeBuildVersion = null;
}
