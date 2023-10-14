# frozen_string_literal: true

require 'bundler/setup'
require 'discordrb'
require 'opus-ruby'
require 'config'
require_relative './core/vcbot'
require_relative './core/voicevox'

# Config
config = Config.load_and_set_settings('./config.yml', './command.yml')

bot = Discordrb::Commands::CommandBot.new(
  token: config.bot.token,
  client_id: config.bot.client_id,
  prefix: config.bot.prefix,
  ignore_bots: true
)

# Bot Init
bot.ready do
  puts 'Bot is Ready!'
  bot.game = config.bot.status
end

bot.command(:summon,
            { aliases: [:s], description: config.command.summon.desc, usage: config.command.summon.usage }) do |event|
  if event.user.voice_channel
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
