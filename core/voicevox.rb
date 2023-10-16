# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

# VoiceVoxとやり取りするためのクラス
class VoiceVox
  def initialize(config)
    @host = config.voicevox.host
  end

  # textとspeakerから音声を生成
  def speak(text, speaker = 1)
    query = voice_query(text, speaker)
    generate_voice(query, speaker)
  end

  private

  # VoiceVoxのクエリを作成
  def voice_query(text, speaker)
    response = post_req('/audio_query', { speaker: speaker, text: text })
    response.body
  end

  # クエリから音声を生成
  def generate_voice(query, speaker)
    response = post_req('/synthesis', { speaker: speaker }, query)
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
