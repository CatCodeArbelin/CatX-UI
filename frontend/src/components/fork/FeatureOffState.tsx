import { Alert, Typography } from 'antd';
import { useTranslation } from 'react-i18next';

export default function FeatureOffState() {
  const { t } = useTranslation();
  return (
    <Alert
      type="info"
      showIcon
      message={t('fork.common.featureOffTitle')}
      description={<Typography.Text>{t('fork.common.featureOffDescription')}</Typography.Text>}
    />
  );
}
