import { Button } from 'antd';
import { SettingOutlined } from '@ant-design/icons';
import { useInRouterContext, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import type { ForkFeature } from '@/lib/fork-feature';
import CatxState from './CatxState';

function FeatureSettingsLink() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  return (
    <Button
      type="link"
      icon={<SettingOutlined />}
      aria-label={t('fork.settings.title')}
      onClick={() => navigate('/settings#catx-features')}
    >
      {t('fork.settings.title')}
    </Button>
  );
}

export default function FeatureOffState({
  feature,
  messageKey,
}: {
  feature?: ForkFeature;
  messageKey?: string;
}) {
  const { t } = useTranslation();
  const inRouter = useInRouterContext();
  return (
    <CatxState
      state="feature_off"
      feature={feature}
      title={messageKey ? t(messageKey) : undefined}
      action={inRouter ? <FeatureSettingsLink /> : undefined}
    />
  );
}
