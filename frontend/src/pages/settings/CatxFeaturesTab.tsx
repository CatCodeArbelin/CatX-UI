import { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, List, Space, Switch, Tag, Typography, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';
import { ForkModuleTitle } from '@/components/fork/ForkModuleMaturity';
import './CatxFeaturesTab.css';

const productionStates = [
  'feature_off',
  'active',
  'initializing',
  'restart_required',
  'error',
] as const;
type FeatureState = (typeof productionStates)[number];
type Feature = {
  key: string;
  enabled: boolean;
  active: boolean;
  state: string;
  requires?: string[];
  restartRequired: boolean;
};
type FeatureResponse = { items: Feature[]; restartRequired: boolean };
const path = '/panel/api/fork/settings/features';

const featureCopy: Record<string, { nameKey: string; detailsKey: string }> = {
  'analytics.enabled': {
    nameKey: 'fork.settings.features.analytics.name',
    detailsKey: 'fork.settings.features.analytics.details',
  },
  'dns_intelligence.enabled': {
    nameKey: 'fork.settings.features.dnsIntelligence.name',
    detailsKey: 'fork.settings.features.dnsIntelligence.details',
  },
  'policies.enabled': {
    nameKey: 'fork.settings.features.policies.name',
    detailsKey: 'fork.settings.features.policies.details',
  },
  'traffic_control.enabled': {
    nameKey: 'fork.settings.features.trafficControl.name',
    detailsKey: 'fork.settings.features.trafficControl.details',
  },
  'security_anomaly.enabled': {
    nameKey: 'fork.settings.features.riskIntelligence.name',
    detailsKey: 'fork.settings.features.riskIntelligence.details',
  },
  'audit.enabled': {
    nameKey: 'fork.settings.features.audit.name',
    detailsKey: 'fork.settings.features.audit.details',
  },
  'self_service.enabled': {
    nameKey: 'fork.settings.features.clientPortal.name',
    detailsKey: 'fork.settings.features.clientPortal.details',
  },
  'fleet_updates.enabled': {
    nameKey: 'fork.settings.features.fleetUpdates.name',
    detailsKey: 'fork.settings.features.fleetUpdates.details',
  },
  'fleet_updates.mutation.enabled': {
    nameKey: 'fork.settings.features.productionFleetUpdates.name',
    detailsKey: 'fork.settings.features.productionFleetUpdates.details',
  },
  'sponsors.enabled': {
    nameKey: 'fork.settings.features.sponsors.name',
    detailsKey: 'fork.settings.features.sponsors.details',
  },
};

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
      setError(t('fork.settings.loadFailed'));
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
  const missingDependencies = (item: Feature) =>
    (item.requires || []).filter((dependency) => !values[dependency]);
  const featureName = (key: string) =>
    t(featureCopy[key]?.nameKey || 'fork.settings.unknownFeature');
  const stateLabel = (state: string) =>
    t(
      productionStates.includes(state as FeatureState)
        ? `fork.settings.states.${state}`
        : 'fork.settings.states.unknown',
    );
  const stateColor = (state: string) =>
    state === 'active'
      ? 'green'
      : state === 'error'
        ? 'red'
        : state === 'initializing'
          ? 'blue'
          : 'gold';
  const dirty = JSON.stringify(values) !== JSON.stringify(savedValues);

  async function save() {
    if (invalidDependencies.length) return;
    setSaving(true);
    const result = await HttpUtil.put<FeatureResponse>(
      path,
      { flags: values },
      { headers: { 'Content-Type': 'application/json' }, silent: true },
    );
    if (!result.success || !result.obj) {
      setError(t('fork.settings.saveFailed'));
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
    <Space
      className="catx-features-tab"
      direction="vertical"
      size="large"
      style={{ width: '100%' }}
    >
      {contextHolder}
      <div>
        <Typography.Title level={2}>
          <ForkModuleTitle moduleId="M00" title={t('fork.settings.title')} />
        </Typography.Title>
        {/* Go-i18n reserves "description" in locale message trees. */}
        <Typography.Paragraph type="secondary">{t('fork.settings.intro')}</Typography.Paragraph>
      </div>
      {items.some((item) => item.restartRequired) && (
        <Alert
          type="warning"
          showIcon
          message={t('fork.settings.restartRequired')}
          description={t('fork.settings.restartRequiredDescription')}
        />
      )}
      {invalidDependencies.length > 0 && (
        <Alert
          type="warning"
          showIcon
          message={t('fork.settings.dependencyWarning')}
          description={
            <Space direction="vertical" size={2}>
              {invalidDependencies.flatMap((item) =>
                missingDependencies(item).map((dependency) => (
                  <Typography.Text key={`${item.key}-${dependency}`}>
                    {t('fork.settings.dependencyRequired', {
                      feature: featureName(item.key),
                      dependency: featureName(dependency),
                    })}
                  </Typography.Text>
                )),
              )}
            </Space>
          }
        />
      )}
      {error && <Alert type="error" showIcon message={error} />}
      <Card loading={loading} size="small">
        <List
          className="catx-features-list"
          dataSource={items}
          locale={{ emptyText: t('fork.settings.noFeatures') }}
          renderItem={(item) => {
            const missing = missingDependencies(item);
            const blocked = missing.length > 0;
            return (
              <List.Item
                className="catx-feature-item"
                actions={[
                  <Space key={item.key} size="small">
                    <Typography.Text type="secondary">
                      {t('fork.settings.desiredState')}
                    </Typography.Text>
                    <Switch
                      aria-label={`${featureName(item.key)}: ${t('fork.settings.desiredState')}`}
                      checked={!!values[item.key]}
                      disabled={blocked}
                      onChange={(enabled) =>
                        setValues((current) => ({ ...current, [item.key]: enabled }))
                      }
                    />
                  </Space>,
                ]}
              >
                <List.Item.Meta
                  title={
                    <Space size="small">
                      <Typography.Text>{featureName(item.key)}</Typography.Text>
                      <Tag color={stateColor(item.state)}>{stateLabel(item.state)}</Tag>
                    </Space>
                  }
                  description={
                    <Space className="catx-feature-description" direction="vertical" size={2}>
                      <Typography.Text type="secondary">
                        {t(
                          featureCopy[item.key]?.detailsKey ||
                            'fork.settings.unknownFeatureDetails',
                        )}
                      </Typography.Text>
                      <Typography.Text type="secondary">
                        {t('fork.settings.runtimeState')}: {stateLabel(item.state)}
                      </Typography.Text>
                      {blocked && (
                        <Tag color="gold">
                          {t('fork.settings.dependencyRequired', {
                            feature: featureName(item.key),
                            dependency: featureName(missing[0]),
                          })}
                        </Tag>
                      )}
                    </Space>
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
