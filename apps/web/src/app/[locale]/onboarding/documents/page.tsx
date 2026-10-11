'use client';

import {useEffect, useState} from 'react';
import {useTranslations} from 'next-intl';
import {useRouter} from '@/i18n/navigation';
import {OnboardingShell} from '@/components/merchant/onboarding-shell';
import {FileDropzone} from '@/components/merchant/file-dropzone';
import {getMe, uploadDocument} from '@/lib/auth-client';
import {useToast} from '@/components/ui/toast-provider';

type DocumentType = 'id_front' | 'id_back' | 'commercial_register';
const types: DocumentType[] = ['id_front', 'id_back', 'commercial_register'];

export default function DocumentsPage({params}: {params: Promise<{locale: string}>}) {
  const [locale, setLocale] = useState('ar');
  const [files, setFiles] = useState<Record<DocumentType, File | null>>({id_front: null, id_back: null, commercial_register: null});
  const [existing, setExisting] = useState<string[]>([]);
  const [rejectedDocument, setRejectedDocument] = useState('');
  const [locked, setLocked] = useState(false);
  const [error, setError] = useState('');
  const t = useTranslations('merchant');
  const router = useRouter();
  const {showToast} = useToast();

  useEffect(() => {
    void params.then((value) => setLocale(value.locale));
    void getMe().then(({data}) => {
      const status = data.merchant.verification_status;
      setLocked(status === 'pending' || status === 'approved' || status === 'suspended');
      setExisting((data.merchant.documents ?? []).map((document) => document.type));
      setRejectedDocument(data.merchant.rejection_document ?? '');
    }).catch(() => setError(t('sessionExpired')));
  }, [params, t]);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (locked) {
      router.push('/onboarding/review');
      return;
    }
    for (const type of types) {
      const file = files[type];
      if (!file && (!existing.includes(type) || type === rejectedDocument)) {
        setError(t('allDocuments'));
        showToast(t('allDocuments'), 'error');
        return;
      }
      if (!file) continue;
      const response = await uploadDocument(type, file);
      if (!response.ok) {
        setError(t('documentError'));
        showToast(t('documentError'), 'error');
        return;
      }
      setExisting((current) => current.includes(type) ? current : [...current, type]);
    }
    showToast(t('saved'), 'success');
    router.push('/onboarding/review');
  };

  return <OnboardingShell locale={locale} step={3}><form className="merchant-form" onSubmit={submit}>
    {locked && <p className="locked-notice">{t('applicationLocked')}</p>}
    {rejectedDocument && <p className="form-error">{t('replaceRejectedDocument', {document: t(`documents.${rejectedDocument}`)})}</p>}
    <p>{t('documentsIntro')}</p>
    {types.map((type) => <FileDropzone
      disabled={locked}
      key={type}
      label={t(`documents.${type}`)}
      hint={existing.includes(type) ? t('documentSaved') : t('documentHint')}
      invalidFile={t('invalidFile')}
      error={type === rejectedDocument && !files[type] ? t('replacementRequired') : undefined}
      file={files[type]}
      onChange={(file) => setFiles((current) => ({...current, [type]: file}))}
    />)}
    {error && <p className="form-error">{error}</p>}
    <button className="button button-primary">{t('next')}</button>
  </form></OnboardingShell>;
}
