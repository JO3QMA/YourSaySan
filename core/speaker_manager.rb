# frozen_string_literal: true

require 'redis'

# 話者設定を管理するクラス
class SpeakerManager
  def initialize(config, logger, voicevox)
    @redis = Redis.new(
      host: config.redis.host,
      port: config.redis.port,
      db: config.redis.db
    )
    @logger = logger
    @default_speaker = 2
    @voicevox = voicevox
  end

  # ユーザーの話者設定を取得
  def get_speaker(user_id)
    speaker = @redis.get("speaker:#{user_id}")
    if speaker
      speaker.to_i
    else
      @default_speaker
    end
  end

  # ユーザーの話者設定を保存
  def set_speaker(user_id, speaker_id)
    @redis.set("speaker:#{user_id}", speaker_id)
    @logger.info('SpeakerManager') { "Set speaker for user #{user_id}: #{speaker_id}" }
  end

  # VoiceVoxのAPIから利用可能な話者の一覧を取得
  def get_available_speakers
    return nil if @voicevox.nil?

    @voicevox.get_speakers
  end

  # 話者IDが有効かチェック
  def valid_speaker?(speaker_id)
    speakers = get_available_speakers
    return false if speakers.nil?

    speakers.key?(speaker_id.to_i)
  end
end
