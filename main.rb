# frozen_string_literal: true

require 'discordrb'
require 'opus-ruby'
require_relative './core/vcbot'
require_relative './core/voicevox'

# Initialize

TOKEN = ''
CLIENT_ID = ''
PREFIX = '!'

bot = Discordrb::Commands::CommandBot.new(
  token: TOKEN,
  client_id: CLIENT_ID,
  prefix: PREFIX,
  ignore_bots: true
)

# 3. ボットの起動時に実行される処理
bot.ready do
  puts 'Botを起動します。'
  bot.game = '[TESTING] 読み上げBot' # ボットのステータスメッセージを設定
end

# help
bot.command(:help, { aliases: [:h] }) do |event|
  message = "
  VOICEVOX読み上げBotです。
  `/s`, `/summon` : 呼び出し
  `/h`, `/help` : ヘルプ
  "
  event.respond(message)
end

bot.command(:summon, { aliases: [:s] }) do |event|
  if event.user.voice_channel
    # if event.user.voice_channel ==
    voice_channel = event.user.voice_channel
    bot.voice_connect(voice_channel)
    event.respond('Hey!')
    vcbot = VCBot.new(bot, event)
    vcbot.main
  else
    event.respond('ボイスチャットに参加してから使用してください。')
  end
end

bot.command(:stop, { aliases: [:skip] }) do |event|
  event.voice.stop_playing if event.voice.playing?
end

bot.command(:bye, { aliases: [:b] }) do |event|
  unless event.message.server.nil?
    bot.voice_destroy(event.message.server)
    event.respond('Bye!')
  end
end

bot.run
