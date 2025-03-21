import {
	Box,
	IconSolidDownload,
	Stack,
	Text,
} from '@highlight-run/ui/components'
import { vars } from '@highlight-run/ui/vars'
import { formatDate } from '@pages/LogsPage/utils'
import { formatNumber } from '@util/numbers'
import moment from 'moment'
import { useMemo } from 'react'

import { HistogramLoading } from '@/pages/Traces/TracesPage'

const LogsCount = ({
	startDate,
	endDate,
	presetSelected,
	totalCount,
	loading,
	onDownload,
}: {
	startDate: Date
	endDate: Date
	presetSelected: boolean
	totalCount: number | undefined
	loading: boolean
	onDownload?: () => void
}) => {
	const dateLabel = useMemo(() => {
		if (presetSelected) {
			return `${formatDate(startDate)} to Now`
		}
		return `${moment(startDate).format('M/D/YY h:mm:ss')} to ${formatDate(
			endDate,
		)}`
	}, [endDate, startDate, presetSelected])

	if (loading) {
		return <HistogramLoading style={{ padding: '6px 0 12px 10px' }} />
	}

	return (
		<Stack
			direction="row"
			gap="8"
			pt="4"
			pb="8"
			px="10"
			align="center"
			style={{ height: 32 }}
		>
			{totalCount !== undefined ? (
				<>
					<Box display="flex" gap="4" flexDirection="row">
						<Text size="xSmall" color="weak">
							{formatNumber(totalCount)} Log
							{totalCount !== 1 ? 's' : ''}
						</Text>
					</Box>
					{onDownload !== undefined ? (
						<IconSolidDownload
							size={12}
							style={{
								color: vars.theme.static.content.weak,
								cursor: 'pointer',
							}}
							onClick={onDownload}
						/>
					) : null}
					<Box br="dividerWeak" style={{ height: 20 }} />
					<Text size="xSmall" color="weak">
						{dateLabel}
					</Text>
				</>
			) : null}
		</Stack>
	)
}

export default LogsCount
