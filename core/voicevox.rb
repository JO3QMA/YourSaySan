# frozen_string_literal: true

require 'net/http'
require 'uri'

# VoiceVoxとやり取りするためのクラス
class VoiceVox
  def initialize(config, logger)
    @host = config.voicevox.host
    @logger = logger
  end

  # textとspeakerから音声を生成
  def speak(text, speaker = 2)
    @logger.infp('VoiceVox') { "Speak: Text: #{text} Speaker: #{speaker}" }
    query = voice_query(text, speaker)
    generate_voice(query, speaker)
  end

  private

  # VoiceVoxのクエリを作成
  def voice_query(text, speaker)
    response = post_req('/audio_query', { speaker: speaker, text: text })
    @logger.debug('VoiceVox') { "Audio_query: Code: #{response.code} Text: #{text}" }
    response.body
  end

  # クエリから音声を生成
  def generate_voice(query, speaker)
    response = post_req('/synthesis', { speaker: speaker }, query)
    @logger.debug('VoiceVox') { "Synthesis: Code: #{response.code}" }
    response.body
  end

  # Post送信用関数
  def post_req(endpoint, query, data = '')
    uri = URI.join(@host, endpoint)
    uri.query = URI.encode_www_form(query)
    header = { 'Content-Type' => 'application/json' }
    Net::HTTP.post(uri, data, header)
  end
end
