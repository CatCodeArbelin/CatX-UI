import { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, List, Space, Switch, Tag, Typography, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';

type Feature = { key: string; enabled: boolean; requires?: string[]; restartRequired: boolean };
type FeatureResponse = { items: Feature[]; restartRequired: boolean };
const path = '/panel/api/fork/settings/features';

export default function CatxFeaturesTab() {
  const { t } = useTranslation();
  const [items, setItems] = useState<Feature[]>([]);
  const [values, setValues] = useState<Record<string, boolean>>({});
  const [savedValues, setSavedValues] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [messageApi, contextHolder] = message.useMessage();

  const load = useCallback(async () => {
    setLoading(true);
    const result = await HttpUtil.get<FeatureResponse>(path, undefined, { silent: true });
    if (!result.success || !result.obj) {
      setError(result.msg || t('fork.settings.loadFailed'));
    } else {
      const next = Object.fromEntries(result.obj.items.map((item) => [item.key, item.enabled]));
      setItems(result.obj.items);
      setValues(next);
      setSavedValues(next);
      setError('');
    }
    setLoading(false);
  }, [t]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  const invalidDependencies = useMemo(
    () =>
      items.filter(
        (item) =>
          values[item.key] && (item.requires || []).some((dependency) => !values[dependency]),
      ),
    [items, values],
  );
  const dirty = JSON.stringify(values) !== JSON.stringify(savedValues);

  async function save() {
    if (invalidDependencies.length) return;
    setSaving(true);
    const result = await HttpUtil.put<FeatureResponse>(path, { flags: values }, { silent: true });
    if (!result.success || !result.obj) {
      setError(result.msg || t('fork.settings.saveFailed'));
    } else {
      const next = Object.fromEntries(result.obj.items.map((item) => [item.key, item.enabled]));
      setItems(result.obj.items);
      setValues(next);
      setSavedValues(next);
      setError('');
      messageApi.success(t('fork.settings.saved'));
    }
    setSaving(false);
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {contextHolder}
      <div>
        <Typography.Title level={2}>{t('fork.settings.title')}</Typography.Title>
        {/* Go-i18n reserves "description" in locale message trees. */}
        <Typography.Paragraph type="secondary">{t('fork.settings.intro')}</Typography.Paragraph>
      </div>
      <Alert
        type="warning"
        showIcon
        message={t('fork.settings.restartRequired')}
        description={t('fork.settings.restartRequiredDescription')}
      />
      {invalidDependencies.length > 0 && (
        <Alert
          type="warning"
          showIcon
          message={t('fork.settings.dependencyWarning')}
          description={invalidDependencies.map((item) => item.key).join(', ')}
        />
      )}
      {error && <Alert type="error" showIcon message={error} />}
      <Card loading={loading} size="small">
        <List
          dataSource={items}
          renderItem={(item) => {
            const blocked = (item.requires || []).some((dependency) => !values[dependency]);
            return (
              <List.Item
                actions={[
                  <Switch
                    key={item.key}
                    checked={!!values[item.key]}
                    disabled={blocked}
                    onChange={(enabled) =>
                      setValues((current) => ({ ...current, [item.key]: enabled }))
                    }
                  />,
                ]}
              >
                <List.Item.Meta
                  title={<Typography.Text code>{item.key}</Typography.Text>}
                  description={
                    blocked ? (
                      <Tag color="gold">{t('fork.settings.dependencyWarning')}</Tag>
                    ) : undefined
                  }
                />
              </List.Item>
            );
          }}
        />
      </Card>
      <Button
        type="primary"
        loading={saving}
        disabled={!dirty || invalidDependencies.length > 0}
        onClick={() => void save()}
      >
        {t('fork.settings.save')}
      </Button>
    </Space>
  );
}
